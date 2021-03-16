/*
Copyright 2020.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package main contains the logic for initialising, configuring, starting and registering the controller.
package main

import (
	"flag"
	"os"
	"strings"

	"github.com/fgiloux/vegeta-operator/operator"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	vegetav1alpha1 "github.com/fgiloux/vegeta-operator/api/v1alpha1"
	"github.com/fgiloux/vegeta-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

type namespaces map[string]struct{}

var (
	cfg     = operator.Config{}
	flagset = flag.CommandLine
	zapOpts = zap.Options{
		Development: true,
	}
)

func printVersion() {
	setupLog.Info("version v1alpha1")
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	flagset.StringVar(&cfg.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&cfg.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flagset.BoolVar(&cfg.EnableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flagset.StringVar(&cfg.Namespaces, "namespaces", "", "Namespaces to scope the interaction of the Vegeta Operator and the apiserver (allow list).")
	flagset.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")
	// Add the zap logger flag set
	zapOpts.BindFlags(flagset)

	// Set up scheme for all resources
	utilruntime.Must(vegetav1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {

	flagset.Parse(os.Args[1:])

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))

	printVersion()

	options := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     cfg.MetricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: cfg.ProbeAddr,
		LeaderElection:         cfg.EnableLeaderElection,
		LeaderElectionID:       "vegeta-2283d09e.testing.io",
	}

	// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	// Note that this is not intended to be used for excluding namespaces, this is better done via a Predicate
	// Also note that you may face performance issues when using this with a high number of namespaces.
	// More Info: https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
	if strings.Contains(cfg.Namespaces, ",") {
		options.Namespace = ""
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(cfg.Namespaces, ","))
	} else {
		options.Namespace = cfg.Namespaces
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	setupLog.Info("manager created")

	if err = (&controllers.VegetaReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Vegeta"),
		Scheme: mgr.GetScheme(),
		Labels: cfg.Labels,
		// TODO: The image should be specified by SHA in the CSV file, which will be injected as environment variable.
		// TODO: I could look at operator conditions (whether I can report operator start failures there, cf OpenShift doc)
		Image: operator.RetrieveDefaultImg(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Vegeta")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
	setupLog.Info("manager started")

	// TODO look at adding results as custom metrics
	// I should have the registration here in init but the details in a separate file
	// similar to what has been done with operator.config
	// https://book.kubebuilder.io/reference/metrics.html#publishing-additional-metrics
}
