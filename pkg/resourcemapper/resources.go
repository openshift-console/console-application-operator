package resourcemapper

import (
	"fmt"

	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
)

const (
	defaultImageStreamNamespace = "openshift"
	imageTriggersAnnotation     = "image.openshift.io/triggers"
)

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
	buildConfigAnnotations := mergeAnnotations(defaultAnnotations(consoleApplication),
		consoleApplication.ObjectMeta.Annotations)

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
							Kind: "ImageStreamTag",
							Name: consoleApplication.Spec.BuildConfiguration.BuilderImage.Image +
								":" + consoleApplication.Spec.BuildConfiguration.BuilderImage.Tag,
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
			AnnotationResolveNames: "*",
			imageTriggersAnnotation: fmt.Sprintf(`[{"from":{"kind":"ImageStreamTag","name":"%v:latest",	"namespace":"%v"},
			"fieldPath":"spec.template.spec.containers[?(@.name==\"%v\")].image","pause":"false"}]`,
				consoleApplication.ObjectMeta.Name, consoleApplication.Namespace, consoleApplication.ObjectMeta.Name),
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
	serviceAnnotations := mergeAnnotations(defaultAnnotations(consoleApplication),
		consoleApplication.ObjectMeta.Annotations)

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
				TargetPort: intstr.FromString(fmt.Sprintf("%d-tcp",
					*consoleApplication.Spec.ResourceConfiguration.Expose.TargetPort)),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
			TLS: &routev1.TLSConfig{Termination: routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect},
		},
	}
}
