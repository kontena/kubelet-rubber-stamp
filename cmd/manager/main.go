package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/kontena/kubelet-rubber-stamp/pkg/apis"
	"github.com/kontena/kubelet-rubber-stamp/pkg/controller"
)

func printVersion() {
	klog.V(2).Infof("Go Version: %s", runtime.Version())
	klog.V(2).Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	klog.V(2).Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	var metricsAddr string

	klog.InitFlags(nil)
	flag.StringVar(&metricsAddr, "metrics-addr", "", fmt.Sprintf("The address the metric endpoint binds to, or \"0\" to disable (default: %s)", metrics.DefaultBindAddress))
	flag.Set("logtostderr", "true")
	flag.Set("v", "2")
	flag.Parse()

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		klog.Fatalf("failed to get watch namespace: %v", err)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Fatal(err)
	}

	switch metricsAddr {
	case "":
		klog.V(2).Infof("Exposing metrics on %s", metrics.DefaultBindAddress)
	case "0":
		klog.V(2).Info("Disabling metrics endpoint")
	default:
		klog.V(2).Infof("Exposing metrics on %s", metricsAddr)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace: namespace,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		klog.Fatal(err)
	}

	klog.V(2).Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		klog.Fatal(err)
	}

	klog.V(2).Info("Starting the Cmd.")

	// Start the Cmd
	klog.Fatal(mgr.Start(signals.SetupSignalHandler()))
}
