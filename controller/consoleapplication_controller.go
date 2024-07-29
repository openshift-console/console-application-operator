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
	operatorCR := &appsv1alpha1.ConsoleApplication{}
	if err := r.Get(ctx, req.NamespacedName, operatorCR); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "ConsoleApplication Not Found")
			return NoRequeue()
		}
		logger.Error(err, "Unable to fetch ConsoleApplication CR")
		SetDegraded(operatorCR, appsv1alpha1.ReasonOperatorResourceNotAvailable.String(), err.Error())
		if err := r.Status().Update(ctx, operatorCR); err != nil {
			return RequeueWithError(err)
		}
		return RequeueOnError(err)
	}

	init := operatorCR.Status.Conditions == nil
	if init {
		operatorCR.Status.Conditions = make([]metav1.Condition, 0)
		SetStarted(operatorCR)
		if err := r.Status().Update(ctx, operatorCR); err != nil {
			return RequeueWithError(err)
		}
	}

	// Fetching the secret resource if specified in the CR
	secretResourceName := operatorCR.Spec.Git.SourceSecretRef
	decodedSecret := ""
	if secretResourceName != "" {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: req.Namespace,
			Name:      secretResourceName,
		}, secret); err != nil {
			logger.Error(err, "Unable To Find Secret Resource")
			SetFailed(operatorCR, appsv1alpha1.ReasonSecretResourceNotFound.String(), err.Error())
			if err := r.Status().Update(ctx, operatorCR); err != nil {
				return RequeueWithError(err)
			}
			return NoRequeue()
		}
		// Decoding the secret data - Considering the Basic Auth secret
		decodedSecret = string(secret.Data["password"])
		logger.Info("Secret Decoded: " + decodedSecret)
	}

	// Checking if the Git Repository is reachable
	gs := gitservice.New(operatorCR.Spec.Git.Url, operatorCR.Spec.Git.Reference, decodedSecret, logger)
	g_status, g_reason := gs.IsRepoReachable()
	logger.Info("Git Repository Reachable: " + string(g_status))

	SetGitServiceCondition(operatorCR, g_status, g_reason.String())
	if err := r.Status().Update(ctx, operatorCR); err != nil {
		return RequeueWithError(err)
	}

	if g_status != metav1.ConditionTrue {
		SetFailed(operatorCR, g_reason.String(), fmt.Sprintf("Git Repository Not Reachable: %s", g_reason.String()))
		if err := r.Status().Update(ctx, operatorCR); err != nil {
			return RequeueWithError(err)
		}
		return NoRequeue()
	}

	// Add the Strategy Service here

	logger.Info("All done!")
	SetSucceeded(operatorCR)
	if err := r.Status().Update(ctx, operatorCR); err != nil {
		return RequeueWithError(err)
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
