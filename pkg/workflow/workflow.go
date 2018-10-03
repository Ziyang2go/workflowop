package operator

import (
	"bytes"
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
	var err error
	switch wfStatus := o.Status.Status; wfStatus {
	case "", "pending":
		err = w.HandlePendingWf(o)
	case "working":
		err = w.HandleWorkingWf(o)
	case "ok", "failed":
		err = w.CleanupWf(o)
	default:
		logrus.Errorf("unknown workflow status %s", wfStatus)
	}
	return err
}

func (w *WorkflowOp) HandlePendingWf(wf *v1alpha.Workflow) error {
	jobs := wf.Inputs.Jobs
	batches := wf.Spec.JobBatch
	statuses := wf.Status.JobStatus
	changed := false
	var updateErr error
	for _, job := range jobs {
		name := job.Name
		batchName := wf.GetObjectMeta().GetName() + "-" + name
		if batches[name] != nil {
			status := wf.Status.JobStatus[name]
			logrus.Printf("%s job %s is in status %s", job.Type, name, status)
			continue
		}
		if batches == nil {
			batches = make(map[string]*v1alpha.BatchReference)
		}
		err := w.CreateJob(job.Type, batchName, job.Data, wf)
		if err != nil {
			logrus.Errorf("failed to create job for %s: %v", name, err)
			continue
		}
		changed = true
		jobName := wf.GetObjectMeta().GetName() + "-" + name
		batches[name] = &v1alpha.BatchReference{"Job", jobName, ""}
		if statuses == nil {
			statuses = make(map[string]string)
		}
		statuses[name] = "working"
	}
	if changed {
		updateErr = w.UpdateWorkflow(batches, statuses, "", wf)
	} else if len(jobs) == len(batches) {
		updateErr = w.UpdateWorkflow(nil, nil, "working", wf)
	}
	if updateErr != nil {
		logrus.Errorf("Update workflow error %v", updateErr)
	}
	return updateErr
}

func (w *WorkflowOp) HandleWorkingWf(wf *v1alpha.Workflow) error {
	statuses := wf.Status.JobStatus
	done := true
	ok := "ok"
	for _, status := range statuses {
		if status != "ok" && status != "failed" {
			// there are jobs not finished
			done = false
			break
		}
		if status == "failed" {
			ok = "failed"
		}
	}
	if done {
		updateErr := w.UpdateWorkflow(nil, nil, ok, wf)
		if updateErr != nil {
			logrus.Errorf("failed to update workflow %v", updateErr)
			return updateErr
		}
	}
	return nil
}

func (w *WorkflowOp) HandleJob(job *batchv1.Job) error {
	logrus.Printf("Handle job %v", job.GetObjectMeta().GetName())
	owner := job.GetOwnerReferences()[0]
	if owner.Kind != "Workflow" {
		return nil
	}
	wfname := owner.Name
	workflow, err := w.GetWorkflowByName(wfname, job.Namespace)
	if err != nil {
		logrus.Errorf("could not get owner reference workflow %v", err)
		return err
	}
	status := workflow.Status.JobStatus[job.Name]
	if status == "ok" || status == "failed" {
		return nil
	}
	finished := job.Status.Succeeded == 1 || job.Status.Failed == 1
	if finished {
		statuses := workflow.Status.JobStatus
		batches := workflow.Spec.JobBatch
		prefix := workflow.Name + "-"
		updateName := strings.TrimPrefix(job.Name, prefix)
		logs := "hello world" //w.GetJobLogs(job)
		if job.Status.Succeeded == 1 {
			statuses[updateName] = "ok"
		} else {
			statuses[updateName] = "failed"
		}
		batches[updateName].Logs = logs
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
	if length := len(jl.Items); length > 1000 {
		return errors.New("job number has limits")
	}
	createJobErr := w.provider.Create(jobTemplate)
	if createJobErr != nil && !kubeerr.IsAlreadyExists(createJobErr) {
		logrus.Errorf("failed to create job for %s: %v", jobName, createJobErr)
		return createJobErr
	}
	return nil
}

func (w *WorkflowOp) GetJobLogs(job *batchv1.Job) string {
	logrus.Println("GET JOBS LOGS .............")

	podName := job.Spec.Template.ObjectMeta.Name
	logrus.Printf("%v", job.Spec.Template.GetObjectMeta())
	logrus.Printf("POD NAME IS .......... %s", podName)
	pod, err := w.GetPodByName(podName, job.Namespace)
	if err != nil {
		logrus.Errorf("failed to get job pod %s", job.Name)
	}
	logs := w.GetLogFromPod(pod)
	return logs
}

func (w *WorkflowOp) GetLogFromPod(pod *corev1.Pod) string {
	client := w.provider.GetKubeClient()
	logOptions := &corev1.PodLogOptions{}
	req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	rc, err := req.Stream()
	if err != nil {
		logrus.Errorf("get log error: %v", err)
	}
	defer rc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	return buf.String()
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

func (w *WorkflowOp) UpdateWorkflow(batchReferences map[string]*v1alpha.BatchReference, batchStatus map[string]string, status string, wf *v1alpha.Workflow) error {
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

func (w *WorkflowOp) GetPodByName(name, namespace string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
	err := w.provider.Get(pod)
	return pod, err
}

func (w *WorkflowOp) CleanupWf(workflow *v1alpha.Workflow) error {
	err := w.provider.Delete(workflow)
	return err
}
