package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/conncache"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime")

type Options struct {
	MetastoreConnector           string
	SystemConnectors             []*runtimev1.Connector
	ConnectionCacheSize          int
	QueryCacheSizeBytes          int64
	SecurityEngineCacheSize      int
	ControllerLogBufferCapacity  int
	ControllerLogBufferSizeBytes int64
	AllowHostAccess              bool
}

type Runtime struct {
	Email          *email.Client
	opts           *Options
	Logger         *zap.Logger
	storage        *storage.Client
	activity       *activity.Client
	metastore      drivers.Handle
	registryCache  *registryCache
	connCache      conncache.Cache
	queryCache     *queryCache
	securityEngine *securityEngine
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, st *storage.Client, ac *activity.Client, emailClient *email.Client) (*Runtime, error) {
	if emailClient == nil {
		emailClient = email.New(email.NewNoopSender())
	}

	rt := &Runtime{
		Email:          emailClient,
		opts:           opts,
		Logger:         logger,
		storage:        st,
		activity:       ac,
		queryCache:     newQueryCache(opts.QueryCacheSizeBytes),
		securityEngine: newSecurityEngine(opts.SecurityEngineCacheSize, logger),
	}

	rt.connCache = rt.newConnectionCache()

	store, _, err := rt.AcquireSystemHandle(ctx, opts.MetastoreConnector)
	if err != nil {
		return nil, err
	}
	rt.metastore = store
	reg, ok := rt.metastore.AsRegistry()
	if !ok {
		return nil, fmt.Errorf("metastore must be a valid registry")
	}

	rt.registryCache = newRegistryCache(rt, reg, logger, ac)
	err = rt.registryCache.init(ctx)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (r *Runtime) AllowHostAccess() bool {
	return r.opts.AllowHostAccess
}

func (r *Runtime) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	r.registryCache.close(ctx)
	err1 := r.queryCache.close()
	err2 := r.connCache.Close(ctx) // Also closes metastore // TODO: Propagate ctx cancellation
	return errors.Join(err1, err2)
}

func (r *Runtime) ResolveSecurity(instanceID string, claims *SecurityClaims, res *runtimev1.Resource) (*ResolvedSecurity, error) {
	inst, err := r.Instance(context.Background(), instanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables(false)
	return r.securityEngine.resolveSecurity(instanceID, inst.Environment, vars, claims, res)
}

// GetInstanceAttributes fetches an instance and converts its annotations to attributes
// nil is returned if an error occurred or instance was not found
func (r *Runtime) GetInstanceAttributes(ctx context.Context, instanceID string) []attribute.KeyValue {
	instance, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil
	}

	return instanceAnnotationsToAttribs(instance)
}

func (r *Runtime) UpdateInstanceWithRillYAML(ctx context.Context, instanceID string, parser *rillv1.Parser, restartController bool) error {
	if parser.RillYAML == nil {
		return errors.New("rill.yaml is required to update an instance")
	}

	rillYAML := parser.RillYAML
	dotEnv := parser.GetDotEnv()

	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	// Shallow clone for editing
	tmp := *inst
	inst = &tmp

	inst.ProjectOLAPConnector = rillYAML.OLAPConnector

	// Dedupe connectors
	connMap := make(map[string]*runtimev1.Connector)
	for _, c := range rillYAML.Connectors {
		connMap[c.Name] = &runtimev1.Connector{
			Type:   c.Type,
			Name:   c.Name,
			Config: c.Defaults,
		}
	}
	for _, r := range parser.Resources {
		if r.ConnectorSpec != nil {
			connMap[r.Name.Name] = &runtimev1.Connector{
				Name:                r.Name.Name,
				Type:                r.ConnectorSpec.Driver,
				Config:              r.ConnectorSpec.Properties,
				TemplatedProperties: r.ConnectorSpec.TemplatedProperties,
				Provision:           r.ConnectorSpec.Provision,
				ProvisionArgs:       r.ConnectorSpec.ProvisionArgs,
				ConfigFromVariables: r.ConnectorSpec.PropertiesFromVariables,
			}
		}
	}

	conns := make([]*runtimev1.Connector, 0, len(connMap))
	for _, c := range connMap {
		conns = append(conns, c)
	}
	inst.ProjectConnectors = conns

	vars := make(map[string]string)
	for _, v := range rillYAML.Variables {
		vars[v.Name] = v.Default
	}
	for k, v := range dotEnv {
		vars[k] = v
	}
	inst.ProjectVariables = vars
	inst.FeatureFlags = rillYAML.FeatureFlags
	inst.PublicPaths = rillYAML.PublicPaths
	return r.EditInstance(ctx, inst, restartController)
}

// UpdateInstanceConnector upserts or removes a connector from an instance
// If connector is nil, the connector is removed; otherwise, it is upserted
func (r *Runtime) UpdateInstanceConnector(ctx context.Context, instanceID, name string, connector *runtimev1.ConnectorSpec) error {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	// remove the connector if it exists
	for i, c := range inst.ProjectConnectors {
		if c.Name == name {
			inst.ProjectConnectors = append(inst.ProjectConnectors[:i], inst.ProjectConnectors[i+1:]...)
			break
		}
	}

	if connector == nil {
		// connector should be removed
		return r.EditInstance(ctx, inst, false)
	}

	// append the new/updated connector
	inst.ProjectConnectors = append(inst.ProjectConnectors, &runtimev1.Connector{
		Name:                name,
		Type:                connector.Driver,
		Config:              connector.Properties,
		TemplatedProperties: connector.TemplatedProperties,
		Provision:           connector.Provision,
		ProvisionArgs:       connector.ProvisionArgs,
		ConfigFromVariables: connector.PropertiesFromVariables,
	})

	return r.EditInstance(ctx, inst, false)
}

func instanceAnnotationsToAttribs(instance *drivers.Instance) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(instance.Annotations)+1)
	attrs = append(attrs, attribute.String("instance_id", instance.ID))
	for k, v := range instance.Annotations {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}
