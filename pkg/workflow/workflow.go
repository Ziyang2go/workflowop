package operator

import (
	"errors"
	"strings"

	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/kube"
	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/v1alpha"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	batches := o.Spec.JobBatch
	statuses := o.Status.JobStatus
	changed := false
	for _, job := range jobs {
		name := job.Name
		batchName := o.GetObjectMeta().GetName() + "-" + name
		if batches[name] != "" {
			status := o.Status.JobStatus[name]
			logrus.Printf("%s job %s is in status %s", job.Type, name, status)
			continue
		}
		logrus.Printf("Creating %s job %s ......", job.Type, name)
		if batches == nil {
			batches = make(map[string]string)
		}
		err := w.CreateJob(job.Type, batchName, job.Data, o)
		if err != nil {
			logrus.Errorf("failed to create job for %s: %v", name, err)
			continue
		}
		changed = true
		batches[name] = o.GetObjectMeta().GetName() + "-" + name
		if statuses == nil {
			statuses = make(map[string]string)
		}
		statuses[name] = "working"
	}

	if !changed {
		shouldUpdate, err := w.shouldUpdateWorkFlow(o)
		if err != nil {
			logrus.Errorf("check workflow error %v", err)
			return err
		}
		if shouldUpdate != "" {
			updateErr := w.UpdateWorkflow(nil, nil, shouldUpdate, o)
			if updateErr != nil {
				logrus.Errorf("Update workflow error %v", updateErr)
			}
			return updateErr
		}
		logrus.Println("no update on workflow")
	} else {
		updateErr := w.UpdateWorkflow(batches, statuses, "", o)
		if updateErr != nil {
			logrus.Errorf("Update workflow error %v", updateErr)
		}
		return updateErr
	}
	return nil
}

func (w *WorkflowOp) shouldUpdateWorkFlow(o *v1alpha.Workflow) (string, error) {
	status := o.Status.Status
	if status == "ok" || status == "failed" {
		//Workflow has already finished
		return "", nil
	}
	if status != "working" {
		return "working", nil
	}
	allFinish := true
	batchStatuses := o.Status.JobStatus
	for _, v := range batchStatuses {
		if v != "ok" && v != "failed" {
			allFinish = false
			break
		}
	}
	if allFinish {
		return "ok", nil
	}
	return "", nil
}

func (w *WorkflowOp) HandleJob(job *batchv1.Job) error {
	logrus.Printf("Handle job %v", job.GetObjectMeta().GetName())
	wfname := job.GetOwnerReferences()[0].Name
	workflow, err := w.GetWorkflowByName(wfname, job.Namespace)
	if err != nil {
		logrus.Errorf("could not get owner reference workflow %v", err)
	}
	status := workflow.Status.JobStatus[job.Name]
	if status == "ok" || status == "failed" {
		return nil
	}
	finished := job.Status.Succeeded == 1 || job.Status.Failed == 1
	if finished {
		statuses := workflow.Status.JobStatus
		prefix := workflow.Name + "-"
		updateName := strings.TrimPrefix(job.Name, prefix)
		if job.Status.Succeeded == 1 {
			statuses[updateName] = "ok"
		} else {
			statuses[updateName] = "failed"
		}
		err := w.UpdateWorkflow(nil, statuses, "", workflow)
		if err != nil {
			logrus.Errorf("Update workflow error %v... ", err)
		}
	}
	return nil
}

func (w *WorkflowOp) CreateJob(jobType, jobName, jobData string, o *v1alpha.Workflow) error {
	jobTemplate := w.GetJobTemplate(jobName, jobName, jobData, o)
	jl, err := w.provider.ListJobs()
	if err != nil {
		logrus.Errorf("failed to list jobs with %v", err)
	}
	logrus.Printf("current  number of jobs is %d ", len(jl.Items))
	if length := len(jl.Items); length > 100 {
		return errors.New("job number has limits")
	}
	createJobErr := w.provider.Create(jobTemplate)
	if createJobErr != nil && !kubeerr.IsAlreadyExists(createJobErr) {
		logrus.Errorf("failed to create job for %s: %v", jobName, createJobErr)
		return createJobErr
	}
	return nil
}

func (w *WorkflowOp) GetWorkflowByName(name, namspace string) (*v1alpha.Workflow, error) {
	workflow := &v1alpha.Workflow{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "threekit.com/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namspace,
		},
	}
	err := w.provider.Get(workflow)
	return workflow, err
}

func (w *WorkflowOp) UpdateWorkflow(batchReferences, batchStatus map[string]string, status string, wf *v1alpha.Workflow) error {
	uploadO := wf.DeepCopy()
	if batchReferences != nil {
		uploadO.Spec.JobBatch = batchReferences
	}
	if batchStatus != nil {
		uploadO.Status.JobStatus = batchStatus
	}
	if status != "" {
		uploadO.Status.Status = status
	}
	updateErr := w.provider.Update(uploadO)
	return updateErr
}

func (w *WorkflowOp) GetJobTemplate(jobType, jobName, jobData string, o *v1alpha.Workflow) *batchv1.Job {
	labels := map[string]string{
		"name": jobName,
	}
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(o, schema.GroupVersionKind{
					Group:   v1alpha.SchemeGroupVersion.Group,
					Version: v1alpha.SchemeGroupVersion.Version,
					Kind:    "Workflow",
				}),
			},
			Labels: labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:    "pi",
							Image:   "perl",
							Command: []string{"perl", "-Mbignum=bpi", "-wle", "print bpi(2000)"},
						},
					},
				},
			},
		},
	}
}
