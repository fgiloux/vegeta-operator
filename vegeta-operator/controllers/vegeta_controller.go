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
	"fmt"
	"reflect"
	"time"

	"github.com/fgiloux/vegeta-operator/operator"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	apiGVStr    = vegetav1alpha1.GroupVersion.String()
)

// +kubebuilder:rbac:groups=vegeta.testing.io,resources=vegeta,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vegeta.testing.io,resources=vegeta/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vegeta.testing.io,resources=vegeta/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *VegetaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vegeta", req.NamespacedName)
	log.V(1).Info("Starting reconciliation")
	// For a Vegeta resource there should be one or many matching pods, that is pods, which have the vegeta object as owner reference.
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
			log.V(1).Info("Vegeta resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, fmt.Errorf("Failed to get Vegeta resource: %v", err)
	}

	statusChanged := false

	// podOwnerKey field is added to the cached pod objects. This key references the owning controller and functions as the index.

	var childPods corev1.PodList

	// Find the list of pods
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingFields{podOwnerKey: req.Name}); err != nil {
		return ctrl.Result{}, fmt.Errorf("List Vegeta's child pods: %v", err)
	}

	// Create new pods if required
	for i := uint(len(childPods.Items)); i < vegeta.Spec.Replicas; i++ {
		pod := r.aPod4Attack(vegeta)
		if err = r.Create(ctx, pod); err != nil {
			log.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return ctrl.Result{}, fmt.Errorf("Failed to create new Pod: %v", err)
		}
		statusChanged = true
		log.V(0).Info("created", "pod", pod)
	}
	if statusChanged {
		// attack pods created, return and requeue
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	// Classify the attack pods
	// PodStatus.phase: Pending, Running, Succeeded, Failed, Unknown
	var activePods []string
	var successfulPods []string
	var failedPods []string

	for i, pod := range childPods.Items {
		if pod.Labels["vegeta.testing.io/type"] == "attack" {
			switch pod.Status.Phase {
			case corev1.PodFailed:
				failedPods = append(failedPods, childPods.Items[i].Name)
			case corev1.PodSucceeded:
				successfulPods = append(successfulPods, childPods.Items[i].Name)
			default:
				activePods = append(activePods, childPods.Items[i].Name)
			}
		}
	}

	applyChanges := func(newPodList *[]string, vegetaPodList *[]string) {
		if !reflect.DeepEqual(*newPodList, *vegetaPodList) {
			*vegetaPodList = *newPodList
			statusChanged = true
		}
	}
	applyChanges(&activePods, &vegeta.Status.Active)
	applyChanges(&successfulPods, &vegeta.Status.Succeeded)
	applyChanges(&failedPods, &vegeta.Status.Failed)
	log.V(1).Info("pod count", "active pods", len(vegeta.Status.Active), "successful pods", len(vegeta.Status.Succeeded), "failed pods", len(vegeta.Status.Failed))

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
		}
		if err := r.Status().Update(ctx, vegeta); err != nil {
			return ctrl.Result{}, fmt.Errorf("Unable to update Vegeta status to reflect pod status: %v", err)
		}
		// status updated with attack pod changes, return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Attack pods have succeeded but report pod may need to be started
	if vegeta.Status.Phase == vegetav1alpha1.SucceededPhase && uint(len(childPods.Items)) < vegeta.Spec.Replicas+1 {
		if vegeta.Spec.Report == nil || vegeta.Spec.Report.OutputType.String() == "" {
			// Nothing to do the report was processed within the attack pod
			vegeta.Status.Phase = vegetav1alpha1.CompletedPhase
			if err := r.Status().Update(ctx, vegeta); err != nil {
				return ctrl.Result{}, fmt.Errorf("Unable to update Vegeta status to completion: %v", err)
			}
			// status updated with completion of report generated by attack pod, return, no need to requeue
			log.V(0).Info("Processing completed", "vegeta", vegeta)
			return ctrl.Result{}, nil
		}
		pod := r.aPod4Report(vegeta)
		if err = r.Create(ctx, pod); err != nil {
			log.Error(err, "Failed to create the pod to generate the report", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return ctrl.Result{}, fmt.Errorf("Failed to create the pod to generate the report: %v", err)
		}
		log.V(0).Info("Report pod created", "pod", pod)
		// Requeue for further processing
		return ctrl.Result{Requeue: true, RequeueAfter: 2 * time.Second}, nil
	}

	/*
			// Refresh Vegeta resource after pod has been created
			err := r.Get(ctx, req.NamespacedName, vegeta)
			if err != nil {
				if errors.IsNotFound(err) {
					// Request object not found, could have been deleted after reconcile request.
					// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
					// Return and don't requeue
					log.V(1).Info("Vegeta resource not found. Ignoring since object must be deleted")
					return ctrl.Result{}, nil
				}
				// Error reading the object - requeue the request.
				return ctrl.Result{}, fmt.Errorf("Failed to get Vegeta resource: %v", err)
			}
			vegeta.Status.Phase = vegetav1alpha1.CompletedPhase
			if err := r.Status().Update(ctx, vegeta); err != nil {
				return ctrl.Result{}, fmt.Errorf("Unable to update Vegeta status to completion: %v", err)
			}
			log.V(0).Info("created", "pod", pod)
			// report poc created, return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	*/

	// Checking the report pod
	if vegeta.Status.Phase == vegetav1alpha1.SucceededPhase && uint(len(childPods.Items)) == vegeta.Spec.Replicas+1 {
		for _, pod := range childPods.Items {
			if pod.Labels["vegeta.testing.io/type"] == "report" {
				switch pod.Status.Phase {
				case corev1.PodFailed:
					vegeta.Status.Phase = vegetav1alpha1.FailedPhase
					if err := r.Status().Update(ctx, vegeta); err != nil {
						return ctrl.Result{}, fmt.Errorf("Unable to update Vegeta status: %v", err)
					}
					// status updated with failure of report pod, return, no need to requeue
					return ctrl.Result{}, nil
				case corev1.PodSucceeded:
					vegeta.Status.Phase = vegetav1alpha1.CompletedPhase
					if err := r.Status().Update(ctx, vegeta); err != nil {
						return ctrl.Result{}, fmt.Errorf("Unable to update Vegeta status: %v", err)
					}
					// status updated with completion of report pod, return, no need to requeue
					return ctrl.Result{}, nil
				default:
					// report pod has not terminated, requeue
					return ctrl.Result{Requeue: true}, nil
				}
			}
		}
	}

	// TODO:
	// It would make sense to implement a validating webhook to force most of the spec fields to be immutable
	// It makes little sense to allow changing them after the pods have been created.
	// https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html

	// Request successfully processed - no requeue
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VegetaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, podOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		pod := rawObj.(*corev1.Pod)
		owner := metav1.GetControllerOf(pod)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Vegeta...
		if owner.APIVersion != apiGVStr || owner.Kind != "Vegeta" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return fmt.Errorf("SetupWithManager: %v", err)
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&vegetav1alpha1.Vegeta{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
