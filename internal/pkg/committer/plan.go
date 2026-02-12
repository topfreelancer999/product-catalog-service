package committer

import (
	"context"

	"github.com/Vektor-AI/commitplan"
	"github.com/Vektor-AI/commitplan/drivers/spanner"
)

// PlanCommitter wraps commitplan.Plan and provides a typed Apply method.
type PlanCommitter struct {
	client *spanner.Client
}

// New creates a new PlanCommitter with the given Spanner client.
func New(client *spanner.Client) *PlanCommitter {
	return &PlanCommitter{client: client}
}

// Apply executes the commit plan atomically.
func (c *PlanCommitter) Apply(ctx context.Context, plan *commitplan.Plan) error {
	if plan == nil {
		return nil
	}
	return plan.Apply(ctx, c.client)
}

