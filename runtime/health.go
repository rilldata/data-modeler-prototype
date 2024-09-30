package runtime

import (
	"context"
	"encoding/json"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type Health struct {
	HangingConn     error
	Registry        error
	InstancesHealth map[string]*InstanceHealth
}

type InstanceHealth struct {
	// always recomputed
	Controller        string
	Repo              string
	ParseErrCount     int
	ReconcileErrCount int

	// cached
	OLAP         string
	MetricsViews map[string]metricsViewHealth

	ControllerVersion int64
}

func (r *Runtime) Health(ctx context.Context) (*Health, error) {
	instances, err := r.registryCache.list()
	if err != nil {
		return nil, err
	}

	ih := make(map[string]*InstanceHealth, len(instances))
	for _, inst := range instances {
		ih[inst.ID], err = r.InstanceHealth(ctx, inst.ID)
		if err != nil && !errors.Is(err, drivers.ErrNotFound) {
			return nil, err
		}
	}
	return &Health{
		HangingConn:     r.connCache.HangingErr(),
		Registry:        r.registryCache.store.(drivers.Handle).Ping(ctx),
		InstancesHealth: ih,
	}, nil
}

func (r *Runtime) InstanceHealth(ctx context.Context, instanceID string) (*InstanceHealth, error) {
	res := &InstanceHealth{}
	// check repo error
	err := r.pingRepo(ctx, instanceID)
	if err != nil {
		res.Repo = err.Error()
	}

	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		res.Controller = err.Error()
		return res, nil
	}

	parser, err := ctrl.Get(ctx, GlobalProjectParserName, false)
	if err != nil {
		return nil, err
	}
	res.ParseErrCount = len(parser.GetProjectParser().State.ParseErrors)

	cachedHealth, _ := r.cachedInstanceHealth(ctx, ctrl.InstanceID, ctrl.catalog.version)
	// set to true if any of the olap engines can be scaled to zero
	var canScaleToZero bool

	// check OLAP error
	olap, release, err := r.OLAP(ctx, instanceID, "")
	if err != nil {
		res.OLAP = err.Error()
	} else {
		mayBeScaledToZero := olap.MayBeScaledToZero(ctx)
		canScaleToZero = canScaleToZero || mayBeScaledToZero
		if cachedHealth != nil && mayBeScaledToZero {
			res.OLAP = cachedHealth.OLAP
		} else {
			err = r.pingOLAP(ctx, olap)
			if err != nil {
				res.OLAP = err.Error()
			}
		}
		release()
	}

	// run queries against metrics views
	resources, err := ctrl.List(ctx, ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}
	res.MetricsViews = make(map[string]metricsViewHealth, len(resources))
	for _, mv := range resources {
		if mv.GetMetricsView().State.ValidSpec == nil {
			if mv.Meta.ReconcileError != "" {
				res.ReconcileErrCount++
			}
			continue
		}
		olap, release, err := r.OLAP(ctx, instanceID, mv.GetMetricsView().State.ValidSpec.Connector)
		if err != nil {
			res.MetricsViews[mv.Meta.Name.Name] = metricsViewHealth{err: err.Error()}
			release()
			continue
		}
		mayBeScaledToZero := olap.MayBeScaledToZero(ctx)
		canScaleToZero = canScaleToZero || mayBeScaledToZero
		release()

		// only use cached health if the OLAP can be scaled to zero
		if cachedHealth != nil && mayBeScaledToZero {
			mvHealth, ok := cachedHealth.MetricsViews[mv.Meta.Name.Name]
			if ok && mvHealth.Version == mv.Meta.StateVersion {
				res.MetricsViews[mv.Meta.Name.Name] = mvHealth
				continue
			}
		}
		_, err = r.Resolve(ctx, &ResolveOptions{
			InstanceID:         ctrl.InstanceID,
			Resolver:           "metricsview_time_range",
			ResolverProperties: map[string]any{"metrics_view": mv.Meta.Name.Name},
			Args:               map[string]any{"priority": 10},
			Claims:             &SecurityClaims{SkipChecks: true},
		})

		mvHealth := metricsViewHealth{Version: mv.Meta.StateVersion}
		if err != nil {
			mvHealth.err = err.Error()
		}
		res.MetricsViews[mv.Meta.Name.Name] = mvHealth
	}

	if !canScaleToZero {
		return res, nil
	}

	// save to cache
	res.ControllerVersion = ctrl.catalog.version
	bytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	err = catalog.UpsertInstanceHealth(ctx, &drivers.InstanceHealth{
		InstanceID: instanceID,
		Health:     bytes,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Runtime) cachedInstanceHealth(ctx context.Context, instanceID string, ctrlVersion int64) (*InstanceHealth, bool) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, false
	}
	defer release()

	cached, err := catalog.FindInstanceHealth(ctx, instanceID)
	if err != nil {
		return nil, false
	}

	c := &InstanceHealth{}
	err = json.Unmarshal(cached.Health, c)
	if err != nil || ctrlVersion != c.ControllerVersion {
		return nil, false
	}
	return c, true
}

func (r *Runtime) pingRepo(ctx context.Context, instanceID string) error {
	repo, rr, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer rr()
	h, ok := repo.(drivers.Handle)
	if !ok {
		return errors.New("unable to ping repo")
	}
	return h.Ping(ctx)
}

func (r *Runtime) pingOLAP(ctx context.Context, olap drivers.OLAPStore) error {
	h, ok := olap.(drivers.Handle)
	if !ok {
		return errors.New("unable to ping olap")
	}
	return h.Ping(ctx)
}

func (h *InstanceHealth) To() *runtimev1.InstanceHealth {
	if h == nil {
		return nil
	}
	r := &runtimev1.InstanceHealth{
		ControllerError:     h.Controller,
		RepoError:           h.Repo,
		OlapError:           h.OLAP,
		ParseErrorCount:     int32(h.ParseErrCount),
		ReconcileErrorCount: int32(h.ReconcileErrCount),
	}
	r.MetricsViewErrors = make(map[string]string, len(h.MetricsViews))
	for k, v := range h.MetricsViews {
		if v.err != "" {
			r.MetricsViewErrors[k] = v.err
		}
	}
	return r
}

type metricsViewHealth struct {
	err     string
	Version int64
}
