package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
	metrics "github.com/openshift-console/console-application-operator/controller/metrics"
	gitservice "github.com/openshift-console/console-application-operator/pkg/git-service"
	"github.com/openshift-console/console-application-operator/pkg/resourcemapper"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	gitSecretKey   string = "password"
	gitSecretValue string
)

// Define a custom result type
type ReconcileResult struct {
	Requeue      bool
	RequeueAfter time.Duration
	Error        error
}

func (r *ConsoleApplicationReconciler) fetchGitSecret(ctx context.Context, req ctrl.Request, consoleApplication *appsv1alpha1.ConsoleApplication, logger logr.Logger) (*corev1.Secret, error) {
	secretResourceName := consoleApplication.Spec.Git.SourceSecretRef
	if secretResourceName == "" {
		return nil, nil // No secret reference specified
	}
	secret, err := r.getResource(ctx, secretResourceName, req.Namespace, &corev1.Secret{}, logger)
	return secret.(*corev1.Secret), err
}

func (r *ConsoleApplicationReconciler) checkGitRepoReachable(req ctrl.Request,
	consoleApplication *appsv1alpha1.ConsoleApplication, decodedSecret string, logger logr.Logger) (status metav1.ConditionStatus, reason gitservice.GitConditionReason) {
	prevStatusCondition := meta.FindStatusCondition(consoleApplication.Status.Conditions, appsv1alpha1.ConditionGitRepoReachable.String())
	if prevStatusCondition != nil {
		logger.Info("Git Repository Reachable Condition Already Exists")
		return prevStatusCondition.Status, gitservice.GitConditionReason(prevStatusCondition.Reason)
	}

	timer := prometheus.NewTimer(metrics.GitRepoReachableDuration.WithLabelValues(req.Namespace, req.Name))
	gs := gitservice.New(consoleApplication.Spec.Git.Url, consoleApplication.Spec.Git.Reference, decodedSecret, logger)
	status, reason = gs.IsRepoReachable(gs)
	timer.ObserveDuration()
	logger.Info("Git Repository Reachable: " + string(status))
	return status, reason
}

func (r *ConsoleApplicationReconciler) handlePostCreationChecks(ctx context.Context, req ctrl.Request, consoleApplication *appsv1alpha1.ConsoleApplication, newObj client.Object, resource resourcemapper.ConsoleAppResource, resMap *resourcemapper.ResourceMapper, logger logr.Logger) ReconcileResult {
	switch obj := newObj.(type) {

	case *buildv1.BuildConfig:
		// Fetch the latest Build of the BuildConfig
		builds, err := r.getResourceList(ctx, &buildv1.BuildList{}, req.Namespace, client.MatchingLabels{"app": consoleApplication.ObjectMeta.Name}, logger)
		if err != nil {
			SetFailed(consoleApplication, resourcemapper.ReasonBuildsNotFound.String(), err.Error())
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return ReconcileResult{Error: err}
			}
			return ReconcileResult{Requeue: false, Error: err}
		}

		requeueNeeded, status, reason, message := resMap.CheckBuildStatus(builds.(*buildv1.BuildList))
		logger.Info(message)
		SetResourceCondition(consoleApplication, resource.ConditionType.String(), status, reason.String(), message)
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return ReconcileResult{Error: err}
		}
		if status == metav1.ConditionFalse || status == metav1.ConditionUnknown {
			if requeueNeeded {
				return ReconcileResult{Requeue: true, RequeueAfter: 30 * time.Second}
			}
			SetFailed(consoleApplication, reason.String(), message)
			return ReconcileResult{Requeue: false, Error: errors.New(message)}
		}

	case *appsv1.Deployment:
		requeueNeeded, status, reason, message := resMap.CheckDeploymentStatus(obj)
		logger.Info(message)
		SetResourceCondition(consoleApplication, resource.ConditionType.String(), status, reason.String(), message)
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return ReconcileResult{Error: err}
		}
		if status == metav1.ConditionFalse || status == metav1.ConditionUnknown {
			if requeueNeeded {
				return ReconcileResult{Requeue: true, RequeueAfter: 10 * time.Second}
			}
			SetFailed(consoleApplication, reason.String(), message)
			return ReconcileResult{Requeue: false, Error: errors.New(message)}
		}

	case *corev1.Service, *routev1.Route:
		if obj == nil {
			logger.Info(fmt.Sprintf("%s Not Ready", resource.Type))
			SetResourceCondition(consoleApplication, resource.ConditionType.String(), metav1.ConditionUnknown, resource.NotReadyReason.String(), fmt.Sprintf("%s Not Ready", resource.Type))
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return ReconcileResult{Error: err}
			}
			return ReconcileResult{Requeue: true, RequeueAfter: 3 * time.Second}
		} else {
			logger.Info(fmt.Sprintf("%s Ready", resource.Type))
			SetResourceCondition(consoleApplication, resource.ConditionType.String(), metav1.ConditionTrue, resource.ReadyReason.String(), fmt.Sprintf("%s Ready", resource.Type))
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return ReconcileResult{Error: err}
			}
		}
		if resource.Type == resourcemapper.Route {
			scheme := "http"
			if obj.(*routev1.Route).Spec.TLS != nil {
				scheme = "https"
			}
			SetStatusField(consoleApplication, appsv1alpha1.StatusFieldApplicationURL,
				fmt.Sprintf("%s://%s", scheme, obj.(*routev1.Route).Spec.Host))
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return ReconcileResult{Error: err}
			}
		}
	}

	return ReconcileResult{Requeue: false}
}

func (r *ConsoleApplicationReconciler) getResource(ctx context.Context, name, namespace string, resourceType client.Object, logger logr.Logger) (client.Object, error) {
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, resourceType)
	if err != nil {
		logger.Error(err, "Unable to find resource", "ResourceType", resourceType, "ResourceName", name)
		return nil, err
	}
	return resourceType, nil
}

func (r *ConsoleApplicationReconciler) getResourceList(ctx context.Context, resourceType client.ObjectList, namespace string, matchLabels client.MatchingLabels, logger logr.Logger) (client.ObjectList, error) {
	var listOptions []client.ListOption

	if namespace != "" {
		listOptions = append(listOptions, client.InNamespace(namespace))
	}
	if matchLabels != nil {
		listOptions = append(listOptions, matchLabels)
	}

	err := r.List(ctx, resourceType, listOptions...)
	if err != nil {
		logger.Error(err, "Unable to list resources", "ResourceType", resourceType)
		return nil, err
	}
	return resourceType, nil
}
