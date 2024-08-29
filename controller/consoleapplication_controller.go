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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gitservice "github.com/openshift-console/console-application-operator/pkg/git-service"
)

// ConsoleApplicationReconciler reconciles a ConsoleApplication object
type ConsoleApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.console.dev,resources=consoleapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.console.dev,resources=consoleapplications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.console.dev,resources=consoleapplications/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ConsoleApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ConsoleApplication...")

	// Retrieving the ConsoleApplication CR instance requested for reconciliation
	consoleApplication := &appsv1alpha1.ConsoleApplication{}
	if err := r.Get(ctx, req.NamespacedName, consoleApplication); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return NoRequeue()
		}
		// Error reading the object - requeue the request.
		return RequeueOnError(err)
	}

	if consoleApplication.Status.Conditions == nil {
		consoleApplication.Status.Conditions = make([]metav1.Condition, 0)
		SetStarted(consoleApplication)
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueOnError(err)
		}
	}

	// Fetching the secret resource if specified in the CR
	secretResourceName := consoleApplication.Spec.Git.SourceSecretRef
	decodedSecret := ""
	if secretResourceName != "" {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: req.Namespace,
			Name:      secretResourceName,
		}, secret); err != nil {
			SetFailed(consoleApplication, appsv1alpha1.ReasonSecretResourceNotFound.String(), err.Error())
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueOnError(err)
			}
			return NoRequeue()
		}
		// Decoding the secret data - Considering the Basic Auth secret
		decodedSecret = string(secret.Data["password"])
	}

	// Checking if the Git Repository is reachable
	gs := gitservice.New(consoleApplication.Spec.Git.Url, consoleApplication.Spec.Git.Reference, decodedSecret, logger)
	gStatus, gReason := gs.IsRepoReachable()
	logger.Info("Git Repository Reachable: " + string(gStatus))

	SetGitServiceCondition(consoleApplication, gStatus, gReason.String())
	if err := r.Status().Update(ctx, consoleApplication); err != nil {
		return RequeueOnError(err)
	}

	if gStatus != metav1.ConditionTrue {
		SetFailed(consoleApplication, gReason.String(), fmt.Sprintf("Git Repository Not Reachable: %s", gReason.String()))
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueOnError(err)
		}
		return NoRequeue()
	}

	// Add the Strategy Service here: Return the list of resources config that needs to be created

	logger.Info("All done!")
	SetSucceeded(consoleApplication)
	if err := r.Status().Update(ctx, consoleApplication); err != nil {
		return RequeueOnError(err)
	}
	return NoRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsoleApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.ConsoleApplication{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
