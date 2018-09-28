package kube

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type Provider interface {
	Create(object runtime.Object) error
	Update(object runtime.Object) error
	Get(object runtime.Object) error
	Delete(object runtime.Object) error
	GetKubeClient() kubernetes.Interface
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
