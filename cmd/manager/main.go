package main

import (
	"flag"
	"os"
	"runtime"

	"github.com/kontena/kubelet-rubber-stamp/pkg/apis"
	"github.com/kontena/kubelet-rubber-stamp/pkg/controller"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const watchNamespaceEnvVar = "WATCH_NAMESPACE"

func printVersion() {
	klog.V(2).Infof("Go Version: %s", runtime.Version())
	klog.V(2).Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("v", "2")
	flag.Parse()

	printVersion()

	namespace, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		klog.Fatalf("failed to get watch namespace: %v", watchNamespaceEnvVar)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
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
