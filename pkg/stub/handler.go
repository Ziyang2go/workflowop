package stub

import (
	"context"

	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/v1alpha"
	operator "github.com/Ziyang2go/workflowop/pkg/workflow"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	batchv1 "k8s.io/api/batch/v1"
)

func NewHandler(workflowop operator.WorkflowOpMethod) sdk.Handler {
	return &Handler{
		operator: workflowop,
	}
}

type Handler struct {
	operator operator.WorkflowOpMethod
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha.Workflow:
		h.operator.HandleWorkflow(o)
	case *batchv1.Job:
		h.operator.HandleJob(o)
	}
	return nil
}
