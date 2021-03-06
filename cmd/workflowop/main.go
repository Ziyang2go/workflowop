package main

import (
	"context"
	"runtime"
	"time"

	"github.com/Ziyang2go/workflowop/pkg/apis/threekit/kube"
	stub "github.com/Ziyang2go/workflowop/pkg/stub"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/Ziyang2go/workflowop/pkg/workflow"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	sdk.ExposeMetricsPort()

	resource := "threekit.com/v1alpha"
	kind := "Workflow"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("failed to get watch namespace: %v", err)
	}
	resyncPeriod := time.Duration(20) * time.Second
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Watch("batch/v1", "Job", namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler(operator.NewWorkflowOp(kube.NewKube())))
	sdk.Run(context.TODO())
}
