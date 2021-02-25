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

// Package controllers contains all the logic for handling vegeta custom resources.
// It implements a reconciliation loop so that test pod(s) get launched and results collected when a new vegeta resource is created.
// Vegeta resources are not expected to be modified.
// TODO: describe parallelisation and how results get retrieved.
package controllers

import (
	"context"
	"reflect"

	"github.com/fgiloux/vegeta-operator/operator"
	"github.com/prometheus/common/log"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	vegetav1alpha1 "github.com/fgiloux/vegeta-operator/api/v1alpha1"
)

// VegetaReconciler reconciles a Vegeta object
type VegetaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Labels operator.Labels
	Image  string
}

var (
	podOwnerKey = ".metadata.controller"
	apiGVStr    = corev1.SchemeGroupVersion.String()
)

// +kubebuilder:rbac:groups=vegeta.testing.io,resources=vegeta,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vegeta.testing.io,resources=vegeta/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vegeta.testing.io,resources=vegeta/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *VegetaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("vegeta", req.NamespacedName)
	reqLogger.Info("Starting reconciliation")
	// TODO(user): Modify the Reconcile function to compare the state specified by
	// the Vegeta object against the actual cluster state, and then
	// perform operations to make the cluster state reflect the state specified by
	// the user.
	//
	// For a Vegeta resource there should be one or many matching pods. TODO: Describe what is used for matching.
	// Once the pod has terminated the Vegeta resource is updated with the result.
	//
	// Using bare pods here rather than jobs. As test executions often require multiple things to be coordinated to get meaningfull results having a mechanism to restart pods when they have not been successful does not bring benefit. In such a scenario it is better to fail fast and allow the user to start again.

	// Fetch the Vegeta instance
	vegeta := &vegetav1alpha1.Vegeta{}
	err := r.Get(ctx, req.NamespacedName, vegeta)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Vegeta resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Vegeta resource")
		return ctrl.Result{}, err
	}

	statusChanged := false

	// TODO: A podOwnerKey field is added to the cached pod objects. This key references the owning controller and functions as the index. I will need to configure the manager to actually index this field.

	var childPods corev1.PodList

	// Find the list of pods
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingFields{podOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child pods")
		return ctrl.Result{}, err
	}

	// Create new pods if required
	for i := uint(len(childPods.Items)); i <= vegeta.Spec.Replicas; i++ {
		pod := r.aPod4Attack(vegeta)
		log.Info("pod created", "pod", pod)
	}

	// Classify the pods
	// PodStatus.phase: Pending, Running, Succeeded, Failed, Unknown
	var activePods []corev1.ObjectReference
	var successfulPods []corev1.ObjectReference
	var failedPods []corev1.ObjectReference

	addPod := func(pod *corev1.Pod, podList []corev1.ObjectReference) {
		podRef, err := ref.GetReference(r.Scheme, pod)
		if err != nil {
			log.Error(err, "unable to make reference to pod", "pod", pod)
			return
		}
		podList = append(podList, *podRef)
	}

	for i, pod := range childPods.Items {
		switch pod.Status.Phase {
		case corev1.PodFailed:
			addPod(&childPods.Items[i], failedPods)
		case corev1.PodSucceeded:
			addPod(&childPods.Items[i], successfulPods)
		default:
			addPod(&childPods.Items[i], activePods)
		}
	}

	applyChanges := func(newPodList []corev1.ObjectReference, vegetaPodList []corev1.ObjectReference) {
		if !reflect.DeepEqual(newPodList, vegetaPodList) {
			vegetaPodList = newPodList
			statusChanged = true
		}
	}
	applyChanges(activePods, vegeta.Status.Active)
	applyChanges(successfulPods, vegeta.Status.Succeeded)
	applyChanges(failedPods, vegeta.Status.Failed)

	log.Debug("pod count", "active pods", len(vegeta.Status.Active), "successful pods", len(vegeta.Status.Succeeded), "failed jobs", len(vegeta.Status.Failed))

	// Update the vegeta status
	if statusChanged {
		if vegeta.Status.Phase != vegetav1alpha1.CompletedPhase && vegeta.Status.Phase != vegetav1alpha1.FailedPhase {
			if len(failedPods) > 0 {
				vegeta.Status.Phase = vegetav1alpha1.FailedPhase
			} else if len(activePods) > 0 {
				vegeta.Status.Phase = vegetav1alpha1.RunningPhase
			} else {
				vegeta.Status.Phase = vegetav1alpha1.SucceededPhase
			}
			// TODO: I need to implement report processing that will bring to the CompletedPhase
		}
		if err := r.Status().Update(ctx, vegeta); err != nil {
			log.Error(err, "unable to update Vegeta status")
			return ctrl.Result{}, err
		}
	}

	// TODO:
	// It would make sense to implement a validating webhook to force most of the spec fields to be immutable
	// It makes little sense to allow changing them after the pods have been created.
	// https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html

	// Request successfully process - no requeue
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VegetaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vegetav1alpha1.Vegeta{}).
		Complete(r)
}
