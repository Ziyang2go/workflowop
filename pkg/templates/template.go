package template

import (
	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/v1alpha"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetJobTemplate(jobType, jobName, jobData string, o *v1alpha.Workflow) *batchv1.Job {
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
