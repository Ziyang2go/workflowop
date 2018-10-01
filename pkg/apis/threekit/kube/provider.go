package kube

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type Provider interface {
	Create(object runtime.Object) error
	Update(object runtime.Object) error
	Get(object runtime.Object) error
	Delete(object runtime.Object) error
	GetKubeClient() kubernetes.Interface
	ListJobs() (*batchv1.JobList, error)
}

type Kube struct {
}

func NewKube() Provider {
	return &Kube{}
}

func (k *Kube) Create(object runtime.Object) error {
	return sdk.Create(object)
}

func (k *Kube) Update(object runtime.Object) error {
	return sdk.Update(object)
}

func (k *Kube) Get(object runtime.Object) error {
	return sdk.Get(object)
}

func (k *Kube) Delete(object runtime.Object) error {
	return sdk.Delete(object)
}

func (k *Kube) GetKubeClient() kubernetes.Interface {
	return k8sclient.GetKubeClient()
}

func (k *Kube) ListJobs() (*batchv1.JobList, error) {
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Printf("failed to get namespace")
	}
	jl := &batchv1.JobList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
	}
	listErr := sdk.List(namespace, jl)
	return jl, listErr
}
