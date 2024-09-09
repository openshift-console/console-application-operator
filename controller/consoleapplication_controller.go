/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"

	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
	metrics "github.com/openshift-console/console-application-operator/controller/metrics"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	gitservice "github.com/openshift-console/console-application-operator/pkg/git-service"
	"github.com/openshift-console/console-application-operator/pkg/resourcemapper"
)

// ConsoleApplicationReconciler reconciles a ConsoleApplication object
type ConsoleApplicationReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	GitServiceInterface gitservice.GitServiceInterface
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsoleApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to the status in the reconciliation loop
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool { return !e.DeleteStateUnknown },
		CreateFunc: func(e event.CreateEvent) bool {
			// Ignore create events for non-ConsoleApplication objects
			_, ok := e.Object.(*appsv1alpha1.ConsoleApplication)
			return ok
		},
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.ConsoleApplication{}).
		Owns(&imagev1.ImageStream{}).
		Owns(&buildv1.BuildConfig{}).
		Owns(&buildv1.Build{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		WithEventFilter(pred).
		Complete(r)
}

//+kubebuilder:rbac:groups=apps.console.dev,resources=consoleapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.console.dev,resources=consoleapplications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.console.dev,resources=consoleapplications/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=build.openshift.io,resources=buildconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=build.openshift.io,resources=builds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ConsoleApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	metrics.ReconcilesTotal.WithLabelValues(req.Namespace, req.Name).Inc()
	logger := log.FromContext(ctx).WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	logger.Info("Reconciling ConsoleApplication...")

	// Retrieving the ConsoleApplication CR instance requested for reconciliation
	consoleApplication, err := r.getResource(ctx, req.Name, req.Namespace, &appsv1alpha1.ConsoleApplication{}, logger)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return NoRequeue()
		}
		// Error reading the object - requeue the request.
		return RequeueOnError(err)
	}

	ca := consoleApplication.(*appsv1alpha1.ConsoleApplication)

	// Set Started Condition
	SetStarted(ca)
	if err := r.Status().Update(ctx, ca); err != nil {
		return RequeueOnError(err)
	}

	// Listing the number of ConsoleApplications in the cluster
	consoleApplications, err := r.getResourceList(ctx, &appsv1alpha1.ConsoleApplicationList{}, "", nil, logger)
	if err != nil {
		logger.Error(err, "Unable to list ConsoleApplications")
		return RequeueOnError(err)
	}
	metrics.ConsoleApplicationsProcessing.With(prometheus.Labels{
		"namespace": req.Namespace,
	}).Set(float64(len(consoleApplications.(*appsv1alpha1.ConsoleApplicationList).Items)))

	// Fetch Secret
	if gitSecret, err := r.fetchGitSecret(ctx, req, ca, logger); err != nil {
		SetFailed(ca, appsv1alpha1.ReasonSecretResourceNotFound.String(), err.Error())
		if err := r.Status().Update(ctx, ca); err != nil {
			return RequeueOnError(err)
		}
		// Stop the reconciliation if the Secret is not found
		return NoRequeue()
	} else if gitSecret != nil {
		gitSecretValue = string(gitSecret.Data[gitSecretKey])
	}

	// Check Git Repository Reachable
	repoReachableStatus, repoReachableReason := r.checkGitRepoReachable(req, ca, gitSecretValue, logger)
	SetGitServiceCondition(ca, repoReachableStatus, repoReachableReason.String())
	if err := r.Status().Update(ctx, ca); err != nil {
		return RequeueOnError(err)
	}

	if repoReachableStatus != metav1.ConditionTrue {
		SetFailed(ca, repoReachableReason.String(), fmt.Sprintf("Git Repository Not Reachable: %s", repoReachableReason.String()))
		if err := r.Status().Update(ctx, ca); err != nil {
			return RequeueOnError(err)
		}
		// Stop the reconciliation if the Git Repository is not reachable
		return NoRequeue()
	}

	// Resource Mapping
	resMap := resourcemapper.NewResourceMapper(ca)
	// Check if the prerequisite resources are available
	requiredResources := resMap.GetRequiredResources()

	for _, reqRes := range requiredResources {
		obj, err := r.getResource(ctx, reqRes.Name, reqRes.Namespace, reqRes.Type, logger)
		if err != nil {
			SetFailed(ca, resourcemapper.ReasonRequiredResourcesNotFound.String(), fmt.Sprintf("Unable to fetch %s (%s) %s", reqRes.Name, reqRes.Type, err.Error()))
			if err := r.Status().Update(ctx, ca); err != nil {
				return RequeueOnError(err)
			}
			// Stop the reconciliation if the required resources are not available
			return NoRequeue()
		}

		if reqRes.CheckStatus != nil {
			ok, err := reqRes.CheckStatus(obj)
			if !ok && err != nil {
				logger.Error(err, "Status Check Failed")
				SetFailed(ca, resourcemapper.ReasonRequiredResourceStatusCheckFailed.String(), err.Error())
				if err := r.Status().Update(ctx, ca); err != nil {
					return RequeueOnError(err)
				}
				return NoRequeue()
			}
		}
	}

	resources, err := resMap.MapResources()
	if err != nil {
		logger.Error(err, "Sanity checks failed")
		SetFailed(ca, appsv1alpha1.ReasonRequirementsNotMet.String(), err.Error())
		if err := r.Status().Update(ctx, ca); err != nil {
			return RequeueOnError(err)
		}
		return NoRequeue()
	}

	for _, resource := range resources {
		newObj := resMap.GetResourceObject(resource.Type)
		// Check if the resource exists
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: req.Namespace,
			Name:      ca.ObjectMeta.Name,
		}, newObj); err != nil {
			if errors.IsNotFound(err) {
				logger.Info(fmt.Sprintf("Creating %s", resource.Type))
				newObj = resource.Object
				_ = ctrl.SetControllerReference(ca, newObj, r.Scheme)
				if err := r.Create(ctx, newObj); err != nil {
					logger.Error(err, fmt.Sprintf("Unable To Create %s", resource.Type))
					SetResourceCondition(ca, resource.ConditionType.String(), metav1.ConditionFalse, resource.CreateFailedReason.String(), err.Error())
					SetFailed(ca, resource.CreateFailedReason.String(), err.Error())
					if err := r.Status().Update(ctx, ca); err != nil {
						return RequeueOnError(err)
					}
					return NoRequeue()
				}
				metrics.ResourcesCreatedTotal.WithLabelValues(req.Namespace, ca.ObjectMeta.Name, string(resource.Type)).Inc()
				logger.Info(fmt.Sprintf("%s Created", resource.Type))
			} else {
				logger.Error(err, fmt.Sprintf("Unable To Fetch %s", resource.Type))
				return NoRequeue()
			}
		}

		// Post creation checks
		result := r.handlePostCreationChecks(ctx, req, ca, newObj, resource, resMap, logger)
		if result.Error != nil {
			if !result.Requeue {
				return NoRequeue()
			}
			return RequeueOnError(result.Error)
		}
		if result.Requeue && result.RequeueAfter > 0 {
			return RequeueAfterSeconds(int(result.RequeueAfter.Seconds()))
		}
	}

	SetSucceeded(ca)
	if err := r.Status().Update(ctx, ca); err != nil {
		return RequeueOnError(err)
	}
	metrics.ConsoleApplicationsSuccessTotal.With(prometheus.Labels{"namespace": req.Namespace}).Inc()
	logger.Info("Reconciliation Successful")
	return NoRequeue()
}
