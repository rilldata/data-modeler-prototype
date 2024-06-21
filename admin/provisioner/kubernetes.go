package provisioner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/c2h5oh/datasize"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"go.uber.org/multierr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	k8serrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

type KubernetesSpec struct {
	Host           string                   `json:"host"`
	Image          string                   `json:"image"`
	Namespace      string                   `json:"namespace"`
	TimeoutSeconds int                      `json:"timeout_seconds"`
	KubeconfigPath string                   `json:"kubeconfig_path"`
	TemplatePaths  *KubernetesTemplatePaths `json:"template_paths"`
}

type KubernetesTemplatePaths struct {
	HTTPIngress string `json:"http_ingress"`
	GRPCIngress string `json:"grpc_ingress"`
	Service     string `json:"service"`
	StatefulSet string `json:"statefulset"`
}

type KubernetesProvisioner struct {
	Spec      *KubernetesSpec
	clientset *kubernetes.Clientset
	templates *template.Template
}

type TemplateData struct {
	Image        string
	ImageTag     string
	Host         string
	CPU          int
	MemoryGB     int
	StorageBytes int64
	Slots        int
	Names        ResourceNames
	Annotations  map[string]string
}

type ResourceNames struct {
	HTTPIngress string
	GRPCIngress string
	Service     string
	StatefulSet string
}

func NewKubernetes(spec json.RawMessage) (*KubernetesProvisioner, error) {
	// Parse the Kubernetes provisioner spec
	ksp := &KubernetesSpec{}
	err := json.Unmarshal(spec, ksp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubernetes provisioner spec: %w", err)
	}

	// Build config from kubeconfig file, this will fallback to in-cluster config if no kubeconfig is specified
	config, err := clientcmd.BuildConfigFromFlags("", ksp.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Create the clientset for the Kubernetes APIs
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Add Sprig template functions (removing functions that leak host info)
	// Derived from Helm: https://github.com/helm/helm/blob/main/pkg/engine/funcs.go
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	// Parse the template definitions
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		ksp.TemplatePaths.HTTPIngress,
		ksp.TemplatePaths.GRPCIngress,
		ksp.TemplatePaths.Service,
		ksp.TemplatePaths.StatefulSet,
	))

	return &KubernetesProvisioner{
		Spec:      ksp,
		clientset: clientset,
		templates: templates,
	}, nil
}

func (p *KubernetesProvisioner) Provision(ctx context.Context, opts *ProvisionOptions) (*Allocation, error) {
	// Get Kubernetes resource names
	names := p.getResourceNames(opts.ProvisionID)

	// Create unique host
	host := p.getHost(opts.ProvisionID)

	// Define template data
	data := &TemplateData{
		ImageTag:     opts.RuntimeVersion,
		Image:        p.Spec.Image,
		Names:        names,
		Host:         strings.Split(host, "//")[1], // Remove protocol
		CPU:          1 * opts.Slots,
		MemoryGB:     2 * opts.Slots,
		StorageBytes: 40 * int64(opts.Slots) * int64(datasize.GB),
		Slots:        opts.Slots,
		Annotations:  opts.Annotations,
	}

	// Define the structured Kubernetes API resources
	httpIng := &netv1.Ingress{}
	grpcIng := &netv1.Ingress{}
	svc := &apiv1.Service{}
	sts := &appsv1.StatefulSet{}

	// Resolve the templates and decode into Kubernetes API resources
	for k, v := range map[string]any{
		p.Spec.TemplatePaths.HTTPIngress: httpIng,
		p.Spec.TemplatePaths.GRPCIngress: grpcIng,
		p.Spec.TemplatePaths.Service:     svc,
		p.Spec.TemplatePaths.StatefulSet: sts,
	} {
		// Resolve template
		buf := &bytes.Buffer{}
		err := p.templates.Lookup(filepath.Base(k)).Execute(buf, data)
		if err != nil {
			return nil, fmt.Errorf("kubernetes provisioner resolve template error: %w", err)
		}

		// Decode into Kubernetes resource
		dec := yaml.NewYAMLOrJSONDecoder(buf, 1000)
		err = dec.Decode(v)
		if err != nil {
			return nil, fmt.Errorf("kubernetes provisioner decode resource error: %w", err)
		}
	}

	// We start by deprovisioning any previous attempt, we do this as a simple way to achieve idempotency
	err := p.Deprovision(ctx, opts.ProvisionID)
	if err != nil {
		return nil, err
	}

	// Create statefulset
	sts.ObjectMeta.Name = names.StatefulSet
	p.addCommonLabels(opts.ProvisionID, sts.ObjectMeta.Labels)
	_, err = p.clientset.AppsV1().StatefulSets(p.Spec.Namespace).Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		err2 := p.Deprovision(ctx, opts.ProvisionID)
		return nil, multierr.Combine(err, err2)
	}

	// Create service
	svc.ObjectMeta.Name = names.Service
	p.addCommonLabels(opts.ProvisionID, svc.ObjectMeta.Labels)
	_, err = p.clientset.CoreV1().Services(p.Spec.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		err2 := p.Deprovision(ctx, opts.ProvisionID)
		return nil, multierr.Combine(err, err2)
	}

	// Create HTTP ingress
	httpIng.ObjectMeta.Name = names.HTTPIngress
	p.addCommonLabels(opts.ProvisionID, httpIng.ObjectMeta.Labels)
	_, err = p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Create(ctx, httpIng, metav1.CreateOptions{})
	if err != nil {
		err2 := p.Deprovision(ctx, opts.ProvisionID)
		return nil, multierr.Combine(err, err2)
	}

	// Create GRPC ingress
	grpcIng.ObjectMeta.Name = names.GRPCIngress
	p.addCommonLabels(opts.ProvisionID, grpcIng.ObjectMeta.Labels)
	_, err = p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Create(ctx, grpcIng, metav1.CreateOptions{})
	if err != nil {
		err2 := p.Deprovision(ctx, opts.ProvisionID)
		return nil, multierr.Combine(err, err2)
	}

	return &Allocation{
		Host:         host,
		Audience:     host,
		CPU:          data.CPU,
		MemoryGB:     data.MemoryGB,
		StorageBytes: data.StorageBytes,
	}, nil
}

func (p *KubernetesProvisioner) Deprovision(ctx context.Context, provisionID string) error {
	// Get Kubernetes resource names
	names := p.getResourceNames(provisionID)

	// Common delete options
	delPolicy := metav1.DeletePropagationForeground
	delOptions := metav1.DeleteOptions{
		PropagationPolicy: &delPolicy,
	}

	// Delete HTTP ingress
	err1 := p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Delete(ctx, names.HTTPIngress, delOptions)

	// Delete GRPC ingress
	err2 := p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Delete(ctx, names.GRPCIngress, delOptions)

	// Delete service
	err3 := p.clientset.CoreV1().Services(p.Spec.Namespace).Delete(ctx, names.Service, delOptions)

	// Delete statefulset
	err4 := p.clientset.AppsV1().StatefulSets(p.Spec.Namespace).Delete(ctx, names.StatefulSet, delOptions)

	// We ignore not found errors for idempotency
	errs := []error{err1, err2, err3, err4}
	for i := 0; i < len(errs); i++ {
		if k8serrs.IsNotFound(errs[i]) {
			errs[i] = nil
		}
	}

	// This returns 'nil' if all errors are 'nil'
	return multierr.Combine(errs...)
}

func (p *KubernetesProvisioner) AwaitReady(ctx context.Context, provisionID string) error {
	// Get Kubernetes resource names
	names := p.getResourceNames(provisionID)

	// Wait for the statefulset to be ready (with timeout)
	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Duration(p.Spec.TimeoutSeconds)*time.Second, true, func(ctx context.Context) (done bool, err error) {
		sts, err := p.clientset.AppsV1().StatefulSets(p.Spec.Namespace).Get(ctx, names.StatefulSet, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return sts.Status.AvailableReplicas > 0 && sts.Status.AvailableReplicas == sts.Status.Replicas && sts.Generation == sts.Status.ObservedGeneration, nil
	})
	if err != nil {
		return err
	}

	// As a final step we make sure the runtime can be reached, we retry on failure, to account for potential small delays in network config propagation
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Logger = nil // Disable inbuilt logger
	pingURL, err := url.JoinPath(p.getHost(provisionID), "/v1/ping")
	if err != nil {
		return err
	}
	resp, err := retryClient.Get(pingURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *KubernetesProvisioner) Update(ctx context.Context, provisionID, newVersion string) error {
	// Get Kubernetes resource names
	names := p.getResourceNames(provisionID)

	// Update the statefulset with retry on conflict to resolve conflicting updates by other clients.
	// More info on this best practice: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version
		sts, err := p.clientset.AppsV1().StatefulSets(p.Spec.Namespace).Get(ctx, names.StatefulSet, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// NOTE: this assumes only one container is defined in the statefulset template
		sts.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", p.Spec.Image, newVersion)

		// Attempt update
		_, err = p.clientset.AppsV1().StatefulSets(p.Spec.Namespace).Update(ctx, sts, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *KubernetesProvisioner) CheckCapacity(ctx context.Context) error {
	// No-op
	return nil
}

func (p *KubernetesProvisioner) getResourceNames(provisionID string) ResourceNames {
	return ResourceNames{
		StatefulSet: fmt.Sprintf("runtime-%s", provisionID),
		Service:     fmt.Sprintf("runtime-%s", provisionID),
		HTTPIngress: fmt.Sprintf("http-runtime-%s", provisionID),
		GRPCIngress: fmt.Sprintf("grpc-runtime-%s", provisionID),
	}
}

func (p *KubernetesProvisioner) getHost(provisionID string) string {
	return strings.ReplaceAll(p.Spec.Host, "*", provisionID)
}

func (p *KubernetesProvisioner) addCommonLabels(provisionID string, resourceLabels map[string]string) {
	resourceLabels["app.kubernetes.io/instance"] = provisionID
	resourceLabels["app.kubernetes.io/managed-by"] = "rill-cloud-admin"
}
