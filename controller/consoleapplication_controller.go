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
	"sort"

	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
	metrics "github.com/openshift-console/console-application-operator/controller/metrics"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	gitservice "github.com/openshift-console/console-application-operator/pkg/git-service"
)

// ConsoleApplicationReconciler reconciles a ConsoleApplication object
type ConsoleApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	gitSecretKey = "password"
)

const (
	defaultImageStreamNamespace = "openshift"
	imageTriggersAnnotation     = "image.openshift.io/triggers"
)

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

	// Listing the number of ConsoleApplications in the cluster
	consoleApplications := &appsv1alpha1.ConsoleApplicationList{}
	if err := r.List(ctx, consoleApplications); err != nil {
		logger.Error(err, "Unable to list ConsoleApplications")
		return RequeueWithError(err)
	}
	metrics.ConsoleApplicationsProcessing.With(prometheus.Labels{
		"namespace": req.Namespace,
	}).Set(float64(len(consoleApplications.Items)))

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
		decodedSecret = string(secret.Data[gitSecretKey])
	}

	if meta.FindStatusCondition(consoleApplication.Status.Conditions,
		appsv1alpha1.ConditionGitRepoReachable.String()) != nil {
		logger.Info("Git Repository Reachable Condition Already Exists")
	} else {
		// Checking if the Git Repository is reachable
		timer := prometheus.NewTimer(metrics.GitRepoReachableDuration.WithLabelValues(req.Namespace, req.Name))
		gs := gitservice.New(consoleApplication.Spec.Git.Url, consoleApplication.Spec.Git.Reference, decodedSecret, logger)
		gStatus, gReason := gs.IsRepoReachable()
		timer.ObserveDuration()
		logger.Info("Git Repository Reachable: " + string(gStatus))

		SetGitServiceCondition(consoleApplication, gStatus, gReason.String())
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueWithError(err)
		}

		if gStatus != metav1.ConditionTrue {
			SetFailed(consoleApplication, gReason.String(), fmt.Sprintf("Git Repository Not Reachable: %s", gReason.String()))
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
			return NoRequeue()
		}
	}

	// Strategy Service

	// Create ImageStream
	imgStream := &imagev1.ImageStream{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      consoleApplication.ObjectMeta.Name,
	}, imgStream); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating ImageStream")
			imgStream = newImageStream(consoleApplication)
			_ = ctrl.SetControllerReference(consoleApplication, imgStream, r.Scheme)
			if err := r.Create(ctx, imgStream); err != nil {
				logger.Error(err, "Unable To Create ImageStream")
				SetBuildConfigCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonImageStreamCreationFailed.String(), err.Error())
				SetFailed(consoleApplication, appsv1alpha1.ReasonImageStreamCreationFailed.String(), err.Error())
				if err := r.Status().Update(ctx, consoleApplication); err != nil {
					return RequeueWithError(err)
				}
				return NoRequeue()
			}
			metrics.ResourcesCreatedTotal.WithLabelValues(req.Namespace, consoleApplication.ObjectMeta.Name, "ImageStream").Inc()
			logger.Info("ImageStream Created")
		} else {
			logger.Error(err, "Unable To Fetch ImageStream")
			return NoRequeue()
		}
	}

	// Check the importStrategy
	if consoleApplication.Spec.ImportStrategy == appsv1alpha1.ImportStrategyBuilderImage {

		if consoleApplication.Spec.BuildConfiguration.BuilderImage.Image == "" ||
			consoleApplication.Spec.BuildConfiguration.BuilderImage.Tag == "" {
			logger.Error(nil, "Builder Image or Tag Not Specified")
			SetFailed(consoleApplication, appsv1alpha1.ReasonRequirementsNotMet.String(), "Builder Image or Tag Not Specified")
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
			return NoRequeue()
		}

		// Check if the ImageStream exists
		imgStream := &imagev1.ImageStream{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: defaultImageStreamNamespace,
			Name:      consoleApplication.Spec.BuildConfiguration.BuilderImage.Image,
		}, imgStream); err != nil {
			logger.Error(err, "Unable To Find ImageStream Resource")
			SetFailed(consoleApplication, appsv1alpha1.ReasonImageStreamNotFound.String(), err.Error())
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
			return NoRequeue()
		}

		// Check if the ImageStream contains the tag
		if ok := checkISTag(imgStream, consoleApplication.Spec.BuildConfiguration.BuilderImage.Tag); !ok {
			logger.Error(nil, "Tag Not Found")
			SetFailed(consoleApplication, appsv1alpha1.ReasonImageStreamNotFound.String(), "Image Stream Tag Not Found")
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
			return NoRequeue()
		}

		if consoleApplication.Spec.BuildConfiguration.BuildOption == appsv1alpha1.BuildOptionBuildConfig {
			// Create BuildConfig
			buildConfig := &buildv1.BuildConfig{}

			// Check if the BuildConfig exists
			if err := r.Get(ctx, client.ObjectKey{
				Namespace: req.Namespace,
				Name:      consoleApplication.ObjectMeta.Name,
			}, buildConfig); err != nil {
				if errors.IsNotFound(err) {
					// Create the BuildConfig
					logger.Info("Creating BuildConfig")

					buildConfig = newBuildConfig(consoleApplication)
					_ = ctrl.SetControllerReference(consoleApplication, buildConfig, r.Scheme)
					if err := r.Create(ctx, buildConfig); err != nil {
						logger.Error(err, "Unable To Create BuildConfig")
						SetBuildConfigCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonBuildConfigCreationFailed.String(), err.Error())
						SetFailed(consoleApplication, appsv1alpha1.ReasonBuildConfigCreationFailed.String(), err.Error())
						if err := r.Status().Update(ctx, consoleApplication); err != nil {
							return RequeueWithError(err)
						}
						return NoRequeue()
					}
					metrics.ResourcesCreatedTotal.WithLabelValues(req.Namespace, consoleApplication.ObjectMeta.Name, "BuildConfig").Inc()
					logger.Info("BuildConfig Created")
					SetBuildConfigCondition(consoleApplication, metav1.ConditionUnknown, appsv1alpha1.ReasonBuildConfigCreated.String(), "BuildConfig Created")
					if err := r.Status().Update(ctx, consoleApplication); err != nil {
						return RequeueWithError(err)
					}
				} else {
					logger.Error(err, "Unable To Fetch BuildConfig")
					return NoRequeue()
				}
			}

			// Fetch the latest Build of the BuildConfig
			builds := &buildv1.BuildList{}
			if err := r.List(ctx, builds, client.InNamespace(req.Namespace), client.MatchingLabels{"app": consoleApplication.ObjectMeta.Name}); err != nil {
				logger.Error(err, "Unable To Fetch Builds")
				SetFailed(consoleApplication, appsv1alpha1.ReasonBuildsNotFound.String(), err.Error())
				if err := r.Status().Update(ctx, consoleApplication); err != nil {
					return RequeueWithError(err)
				}
				return NoRequeue()
			}

			if len(builds.Items) == 0 {
				logger.Info("No Builds Found")
				SetBuildConfigCondition(consoleApplication, metav1.ConditionUnknown, appsv1alpha1.ReasonBuildsNotFound.String(), "Builds not created yet")
				if err := r.Status().Update(ctx, consoleApplication); err != nil {
					return RequeueWithError(err)
				}
				return RequeueAfterSeconds(30)
			}

			if len(builds.Items) > 0 {
				// Sort the builds by creation timestamp like the latest build will be the first element
				sort.Slice(builds.Items, func(i, j int) bool {
					return builds.Items[i].CreationTimestamp.After(builds.Items[j].CreationTimestamp.Time)
				})

				// Fetch the latest build
				latestBuild := builds.Items[0]
				logger.Info("Latest Build: " + latestBuild.Name + ", Latest Build Status: " + string(latestBuild.Status.Phase))

				if latestBuild.Status.Phase == buildv1.BuildPhaseNew || latestBuild.Status.Phase == buildv1.BuildPhasePending || latestBuild.Status.Phase == buildv1.BuildPhaseRunning {
					logger.Info("Build In Progress: " + string(latestBuild.Status.Phase))
					SetBuildConfigCondition(consoleApplication, metav1.ConditionUnknown, string(latestBuild.Status.Phase), "Build In Progress")
					if err := r.Status().Update(ctx, consoleApplication); err != nil {
						return RequeueWithError(err)
					}
					return RequeueAfterSeconds(30)
				}

				if latestBuild.Status.Phase == buildv1.BuildPhaseComplete {
					logger.Info("Build Completed: " + string(latestBuild.Status.Phase))
					SetBuildConfigCondition(consoleApplication, metav1.ConditionTrue, string(latestBuild.Status.Phase), "Build Completed")
					if err := r.Status().Update(ctx, consoleApplication); err != nil {
						return RequeueWithError(err)
					}
				} else {
					logger.Info("Build Failed: " + string(latestBuild.Status.Phase))
					SetBuildConfigCondition(consoleApplication, metav1.ConditionFalse, string(latestBuild.Status.Phase), "Build Failed")
					SetFailed(consoleApplication, appsv1alpha1.ReasonBuildsFailed.String(), "Build Failed")
					if err := r.Status().Update(ctx, consoleApplication); err != nil {
						return RequeueWithError(err)
					}
					return NoRequeue()
				}
			}
		}
	}

	// Create Workload - Deployment
	workload := &appsv1.Deployment{}
	// Check if the Deployment exists
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      consoleApplication.ObjectMeta.Name,
	}, workload); err != nil {
		if errors.IsNotFound(err) {
			// Create the Deployment
			logger.Info("Creating K8s Deployment")
			workload = newK8sDeployment(consoleApplication)
			_ = ctrl.SetControllerReference(consoleApplication, workload, r.Scheme)
			if err := r.Create(ctx, workload); err != nil {
				logger.Error(err, "Unable To Create Deployment")
				SetWorkloadCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonWorkloadCreationFailed.String(), err.Error())
				SetFailed(consoleApplication, appsv1alpha1.ReasonWorkloadCreationFailed.String(), err.Error())
				if err := r.Status().Update(ctx, consoleApplication); err != nil {
					return RequeueWithError(err)
				}
				return NoRequeue()
			}
			metrics.ResourcesCreatedTotal.WithLabelValues(req.Namespace, consoleApplication.ObjectMeta.Name, "Deployment").Inc()
			logger.Info("Workload Created")
			return RequeueAfterSeconds(3)
		} else {
			logger.Error(err, "Unable To Fetch Workload")
			return NoRequeue()
		}
	}

	if workload.Status.Conditions == nil {
		logger.Info("Deployment Not Found")
		SetWorkloadCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonWorkloadNotReady.String(), "Deployment Not Found")
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueWithError(err)
		}
		return RequeueAfterSeconds(3)
	}

	// Check the status of the Workload
	availableStatusCondition := appsv1.DeploymentCondition{}
	progressingStatusCondition := appsv1.DeploymentCondition{}
	for _, condition := range workload.Status.Conditions {
		switch condition.Type {
		case appsv1.DeploymentAvailable:
			availableStatusCondition = condition
		case appsv1.DeploymentProgressing:
			progressingStatusCondition = condition
		}
	}

	if progressingStatusCondition.Status == corev1.ConditionTrue {
		if availableStatusCondition.Status == corev1.ConditionFalse {
			logger.Info("Deployment Not Ready")
			SetWorkloadCondition(consoleApplication, metav1.ConditionUnknown, appsv1alpha1.ReasonWorkloadNotReady.String(), "Deployment Not Ready")
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
			return RequeueAfterSeconds(10)
		} else if availableStatusCondition.Status == corev1.ConditionTrue {
			// Deployment is ready
			logger.Info("Deployment Ready")
			SetWorkloadCondition(consoleApplication, metav1.ConditionTrue, appsv1alpha1.ReasonWorkloadReady.String(), "Deployment Ready")
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
		}
	} else {
		logger.Info("Deployment Progressing Failed")
		SetWorkloadCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonWorkloadCreationFailed.String(), "Deployment Progressing Failed")
		SetFailed(consoleApplication, appsv1alpha1.ReasonWorkloadCreationFailed.String(), "Deployment Progressing Failed")
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueOnError(err)
		}
		return NoRequeue()
	}

	// Create Service
	service := &corev1.Service{}
	// Check if the Service exists
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      consoleApplication.ObjectMeta.Name,
	}, service); err != nil {
		if errors.IsNotFound(err) {
			// Create the Service
			logger.Info("Creating K8s Service")
			service = newService(consoleApplication)
			_ = ctrl.SetControllerReference(consoleApplication, service, r.Scheme)
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "Unable To Create Service")
				SetServiceCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonServiceCreationFailed.String(), err.Error())
				SetFailed(consoleApplication, appsv1alpha1.ReasonServiceCreationFailed.String(), err.Error())
				if err := r.Status().Update(ctx, consoleApplication); err != nil {
					return RequeueWithError(err)
				}
				return NoRequeue()
			}
			metrics.ResourcesCreatedTotal.WithLabelValues(req.Namespace, consoleApplication.ObjectMeta.Name, "Service").Inc()
			logger.Info("Service Created")
		} else {
			logger.Error(err, "Unable To Fetch Service")
			return NoRequeue()
		}
	}

	// Check if Service is ready
	if service == nil {
		logger.Info("Service Not Ready")
		SetServiceCondition(consoleApplication, metav1.ConditionUnknown, appsv1alpha1.ReasonServiceNotReady.String(), "Service Not Ready")
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueWithError(err)
		}
		return RequeueAfterSeconds(3)
	} else {
		logger.Info("Service Ready")
		SetServiceCondition(consoleApplication, metav1.ConditionTrue, appsv1alpha1.ReasonServiceReady.String(), "Service Ready")
		if err := r.Status().Update(ctx, consoleApplication); err != nil {
			return RequeueWithError(err)
		}
	}

	// Create Route if mentioned in the CR
	if consoleApplication.Spec.ResourceConfiguration.Expose.CreateRoute {
		route := &routev1.Route{}
		// Check if the Route exists
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: req.Namespace,
			Name:      consoleApplication.ObjectMeta.Name,
		}, route); err != nil {
			if errors.IsNotFound(err) {
				// Create the Route
				logger.Info("Creating Route")
				route = newRoute(consoleApplication)
				_ = ctrl.SetControllerReference(consoleApplication, route, r.Scheme)
				if err := r.Create(ctx, route); err != nil {
					logger.Error(err, "Unable To Create Route")
					SetRouteCondition(consoleApplication, metav1.ConditionFalse, appsv1alpha1.ReasonRouteCreationFailed.String(), err.Error())
					SetFailed(consoleApplication, appsv1alpha1.ReasonRouteCreationFailed.String(), err.Error())
					if err := r.Status().Update(ctx, consoleApplication); err != nil {
						return RequeueWithError(err)
					}
					return NoRequeue()
				}
				metrics.ResourcesCreatedTotal.WithLabelValues(req.Namespace, consoleApplication.ObjectMeta.Name, "Route").Inc()
				logger.Info("Route Created")
			} else {
				logger.Error(err, "Unable To Fetch Route")
				return NoRequeue()
			}
		}

		// Check if Route is ready
		if route == nil {
			logger.Info("Route Not Ready")
			SetRouteCondition(consoleApplication, metav1.ConditionUnknown, appsv1alpha1.ReasonRouteNotReady.String(), "Route Not Ready")
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
			return RequeueAfterSeconds(3)
		} else {
			logger.Info("Route Ready")
			SetRouteCondition(consoleApplication, metav1.ConditionTrue, appsv1alpha1.ReasonRouteReady.String(), "Route Ready")
			SetRouteURL(consoleApplication, route)
			if err := r.Status().Update(ctx, consoleApplication); err != nil {
				return RequeueWithError(err)
			}
		}
	}

	SetSucceeded(consoleApplication)
	if err := r.Status().Update(ctx, consoleApplication); err != nil {
		return RequeueOnError(err)
	}
	metrics.ConsoleApplicationsSuccessTotal.With(prometheus.Labels{"namespace": req.Namespace}).Inc()
	logger.Info("All done!")
	return NoRequeue()
}

func checkISTag(imgStream *imagev1.ImageStream, tag string) bool {
	for _, tagEvent := range imgStream.Status.Tags {
		if tagEvent.Tag == tag {
			return true
		}
	}
	return false
}

func newImageStream(consoleApplication *appsv1alpha1.ConsoleApplication) *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consoleApplication.ObjectMeta.Name,
			Namespace:   consoleApplication.Namespace,
			Labels:      defaultLabels(consoleApplication),
			Annotations: defaultAnnotations(consoleApplication),
		},
	}
}

func newBuildConfig(consoleApplication *appsv1alpha1.ConsoleApplication) *buildv1.BuildConfig {

	buildConfigLabels := mergeLabels(defaultLabels(consoleApplication), consoleApplication.ObjectMeta.Labels)
	buildConfigAnnotations := mergeAnnotations(defaultAnnotations(consoleApplication), consoleApplication.ObjectMeta.Annotations)

	return &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consoleApplication.ObjectMeta.Name,
			Namespace:   consoleApplication.ObjectMeta.Namespace,
			Labels:      buildConfigLabels,
			Annotations: buildConfigAnnotations,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Source: buildv1.BuildSource{
					ContextDir: consoleApplication.Spec.Git.ContextDir,
					Type:       buildv1.BuildSourceGit,
					Git: &buildv1.GitBuildSource{
						URI: consoleApplication.Spec.Git.Url,
						Ref: consoleApplication.Spec.Git.Reference,
					},
					SourceSecret: func() *corev1.LocalObjectReference {
						if consoleApplication.Spec.Git.SourceSecretRef != "" {
							return &corev1.LocalObjectReference{
								Name: consoleApplication.Spec.Git.SourceSecretRef,
							}
						}
						return nil
					}(),
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.SourceBuildStrategyType,
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      consoleApplication.Spec.BuildConfiguration.BuilderImage.Image + ":" + consoleApplication.Spec.BuildConfiguration.BuilderImage.Tag,
							Namespace: defaultImageStreamNamespace,
						},
						Env: consoleApplication.Spec.BuildConfiguration.Env,
					},
				},
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: consoleApplication.ObjectMeta.Name + ":latest",
					},
				},
				NodeSelector: nil,
			},
			Triggers: []buildv1.BuildTriggerPolicy{
				{Type: buildv1.ConfigChangeBuildTriggerType},
				{Type: buildv1.ImageChangeBuildTriggerType,
					ImageChange: &buildv1.ImageChangeTrigger{}},
				{Type: buildv1.GenericWebHookBuildTriggerType,
					GenericWebHook: &buildv1.WebHookTrigger{
						Secret: consoleApplication.ObjectMeta.Name + "-generic-webhook-secret",
					},
				},
			},
		},
	}
}

func newK8sDeployment(consoleApplication *appsv1alpha1.ConsoleApplication) *appsv1.Deployment {
	deploymentLabels := mergeLabels(defaultLabels(consoleApplication), consoleApplication.ObjectMeta.Labels)

	deploymentAnnotations := mergeAnnotations(
		defaultAnnotations(consoleApplication),
		consoleApplication.ObjectMeta.Annotations,
		map[string]string{
			AnnotationResolveNames:  "*",
			imageTriggersAnnotation: fmt.Sprintf(`[{"from":{"kind":"ImageStreamTag","name":"%v:latest","namespace":"%v"},"fieldPath":"spec.template.spec.containers[?(@.name==\"%v\")].image","pause":"false"}]`, consoleApplication.ObjectMeta.Name, consoleApplication.Namespace, consoleApplication.ObjectMeta.Name),
		},
	)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consoleApplication.ObjectMeta.Name,
			Namespace:   consoleApplication.Namespace,
			Labels:      deploymentLabels,
			Annotations: deploymentAnnotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: consoleApplication.Spec.ResourceConfiguration.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": consoleApplication.ObjectMeta.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: mergeLabels(
						consoleApplication.ObjectMeta.Labels,
						map[string]string{
							"app": consoleApplication.ObjectMeta.Name,
						}),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: consoleApplication.ObjectMeta.Name,
							Image: "image-registry.openshift-image-registry.svc:5000/" +
								consoleApplication.ObjectMeta.Namespace + "/" + consoleApplication.ObjectMeta.Name + ":latest",
							Env: consoleApplication.Spec.ResourceConfiguration.Env,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: *consoleApplication.Spec.ResourceConfiguration.Expose.TargetPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(consoleApplication *appsv1alpha1.ConsoleApplication) *corev1.Service {
	serviceLabels := mergeLabels(defaultLabels(consoleApplication), consoleApplication.ObjectMeta.Labels)
	serviceAnnotations := mergeAnnotations(defaultAnnotations(consoleApplication), consoleApplication.ObjectMeta.Annotations)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consoleApplication.ObjectMeta.Name,
			Namespace:   consoleApplication.Namespace,
			Labels:      serviceLabels,
			Annotations: serviceAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": consoleApplication.ObjectMeta.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       fmt.Sprintf("%d-tcp", *consoleApplication.Spec.ResourceConfiguration.Expose.TargetPort),
					Port:       *consoleApplication.Spec.ResourceConfiguration.Expose.TargetPort,
					TargetPort: intstr.FromInt(int(*consoleApplication.Spec.ResourceConfiguration.Expose.TargetPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func newRoute(consoleApplication *appsv1alpha1.ConsoleApplication) *routev1.Route {
	routeLabels := mergeLabels(defaultLabels(consoleApplication), consoleApplication.ObjectMeta.Labels)
	routeAnnotations := mergeAnnotations(
		defaultAnnotations(consoleApplication),
		consoleApplication.ObjectMeta.Annotations,
		map[string]string{AnnotationHostGenerated: "true"},
	)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consoleApplication.ObjectMeta.Name,
			Namespace:   consoleApplication.Namespace,
			Labels:      routeLabels,
			Annotations: routeAnnotations,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: consoleApplication.ObjectMeta.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(fmt.Sprintf("%d-tcp", *consoleApplication.Spec.ResourceConfiguration.Expose.TargetPort)),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
			TLS: &routev1.TLSConfig{Termination: routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect},
		},
	}
}
