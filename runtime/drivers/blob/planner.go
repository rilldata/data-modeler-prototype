package blob

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/container"
	"gocloud.dev/blob"
)

// planner keeps items as per extract policy
// it adds objects in the container which stops consuming files once it reaches file extract policy limits
// every objects has details about what is the download strategy for that object
type planner struct {
	policy *runtimev1.Source_ExtractPolicy
	// rowPlanner adds support for row extract policy
	rowPlanner rowPlanner
	// keeps collection of objects to be downloaded
	// also adds support for file extract policy
	container container.Container[*objectWithPlan]
}

func newPlanner(policy *runtimev1.Source_ExtractPolicy) (*planner, error) {
	c, err := containerForFileStrategy(policy)
	if err != nil {
		return nil, err
	}

	rowPlanner := rowPlannerForRowStrategy(policy)
	return &planner{
		policy:     policy,
		container:  c,
		rowPlanner: rowPlanner,
	}, nil
}

func (p *planner) add(item *blob.ListObject) bool {
	if p.done() {
		return false
	}

	obj := p.rowPlanner.planFile(item)
	return p.container.Add(obj)
}

func (p *planner) done() bool {
	return p.container.Full() || p.rowPlanner.done()
}

func (p *planner) items() []*objectWithPlan {
	return p.container.Items()
}

func containerForFileStrategy(policy *runtimev1.Source_ExtractPolicy) (container.Container[*objectWithPlan], error) {
	strategy := runtimev1.Source_ExtractPolicy_STRATEGY_UNSPECIFIED
	limit := 0
	if policy != nil {
		strategy = policy.FilesStrategy
		limit = int(policy.FilesLimit)
	}

	switch strategy {
	case runtimev1.Source_ExtractPolicy_STRATEGY_TAIL:
		return container.NewFIFO[*objectWithPlan](limit, nil)
	case runtimev1.Source_ExtractPolicy_STRATEGY_HEAD:
		return container.NewBounded[*objectWithPlan](limit)
	default:
		// No option selected
		return container.NewUnbounded[*objectWithPlan]()
	}
}

func rowPlannerForRowStrategy(policy *runtimev1.Source_ExtractPolicy) rowPlanner {
	if policy == nil {
		return &plannerWithoutLimits{}
	}

	if policy.RowsStrategy != runtimev1.Source_ExtractPolicy_STRATEGY_UNSPECIFIED {
		if policy.FilesStrategy != runtimev1.Source_ExtractPolicy_STRATEGY_UNSPECIFIED {
			// file strategy specified row limits are per file
			return &plannerWithPerFileLimits{strategy: policy.RowsStrategy, limitInBytes: policy.RowsLimitBytes}
		}
		// global policy since file strategy is not specified
		return &plannerWithGlobalLimits{strategy: policy.RowsStrategy, limitInBytes: policy.RowsLimitBytes}
	}
	return &plannerWithoutLimits{}
}
