package stub

import (
	"context"

	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/v1alpha"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
)

func NewHandler(workflowop) sdk.Handler {
	return &Handler{
		workflowop: workflowop,
	}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha.Workflow:
		logrus.Printf("Workflow update %v", o)
	case *batchv1.Job:
		logrus.Printf("Job Update %v", o)
	}
	return nil
}
