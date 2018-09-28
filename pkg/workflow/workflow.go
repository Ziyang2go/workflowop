package operator

import (
	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/kube"
	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/v1alpha"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
)

type WorkflowOpMethod interface {
	HandleWorkflow(*v1alpha.Workflow) error
	HandleJob(*batchv1.Job) error
}

type WorkflowOp struct {
	provider kube.Provider
}

func NewWorkflowOp(provider kube.Provider) WorkflowOpMethod {
	return &WorkflowOp{
		provider: provider,
	}
}

func (w *WorkflowOp) HandleWorkflow(o *v1alpha.Workflow) error {
	jobs := o.Inputs.Jobs
	for _, job := range jobs {
		logrus.Printf("job type is %s", job.Type)
		logrus.Printf("job name is %s", job.Name)
	}
	return nil
}

func (w *WorkflowOp) HandleJob(job *batchv1.Job) error {
	logrus.Printf("Handle job %v", job)
	if job.Status.Succeeded == 1 || job.Status.Failed == 1 {
		logrus.Println("Job completed ")
	}
	return nil
}

func (w *WorkflowOp) UpdateWorkflow(spec, status string) error {
	logrus.Printf("Update workflow %s %s", spec, status)
	return nil
}

func (w *WorkflowOp) CheckJobs() int {
	return 10
}

func (w *WorkflowOp) getJob(jobType string) error {
	logrus.Printf("Work up .........%v", jobType)
	return nil
}

func (w *WorkflowOp) UpdateJob() error {
	return nil
}

func (w *WorkflowOp) CheckStatus(o *v1alpha.Workflow) error {
	logrus.Printf("Check workflow status......%v", o)
	return nil
}
