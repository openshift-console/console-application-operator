/*
Package resourcemapper provides the list of resources
that needs to be created for the consoleApplication
*/

package resourcemapper

import (
	"fmt"
	"sort"

	appsv1alpha1 "github.com/openshift-console/console-application-operator/api/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceMapper struct {
	consoleApplication *appsv1alpha1.ConsoleApplication
	Resources          []ConsoleAppResource
}

type ConsoleAppResource struct {
	Type               ResourceType
	Object             client.Object
	ConditionType      ConditionType
	CreateFailedReason ConditionReason
	NotReadyReason     ConditionReason
	ReadyReason        ConditionReason
}

type RequiredResource struct {
	Name        string
	Namespace   string
	Type        client.Object
	CheckStatus func(client.Object) (bool, error)
}

type ResourceType string

const (
	ImageStream ResourceType = "ImageStream"
	BuildConfig ResourceType = "BuildConfig"
	Deployment  ResourceType = "Deployment"
	Service     ResourceType = "Service"
	Route       ResourceType = "Route"
)

func NewResourceMapper(consoleApplication *appsv1alpha1.ConsoleApplication) *ResourceMapper {
	return &ResourceMapper{
		consoleApplication: consoleApplication,
	}
}

// RequiredResources returns the list of resources that are required according to the ConsoleApplication spec.
func (r *ResourceMapper) GetRequiredResources() (resources []RequiredResource) {
	if r.consoleApplication.Spec.ImportStrategy == appsv1alpha1.ImportStrategyBuilderImage {
		// Check if the ImageStream for the BuilderImage exists
		resources = append(resources, RequiredResource{
			Name:      r.consoleApplication.Spec.BuildConfiguration.BuilderImage.Image,
			Namespace: defaultImageStreamNamespace,
			Type:      &imagev1.ImageStream{},
			CheckStatus: func(res client.Object) (bool, error) {
				return r.checkBuilderImageTagExists(res.(*imagev1.ImageStream), r.consoleApplication.Spec.BuildConfiguration.BuilderImage.Tag)
			},
		})
	}
	return resources
}

func (r *ResourceMapper) MapResources() ([]ConsoleAppResource, error) {

	isBuilderImageImportStrategy := r.consoleApplication.Spec.ImportStrategy == appsv1alpha1.ImportStrategyBuilderImage
	isBuildOptionBuildConfig := r.consoleApplication.Spec.BuildConfiguration.BuildOption == appsv1alpha1.BuildOptionBuildConfig
	isDeploymentResourceType := r.consoleApplication.Spec.ResourceConfiguration.ResourceType == appsv1alpha1.ResourceTypeDeployment
	isRouteCreationEnabled := r.consoleApplication.Spec.ResourceConfiguration.Expose.CreateRoute

	// Perform sanity checks
	if err := r.sanityCheck(); err != nil {
		return nil, err
	}

	// Create ImageStream
	r.Resources = append(r.Resources, ConsoleAppResource{
		Type:               ImageStream,
		Object:             newImageStream(r.consoleApplication),
		ConditionType:      ConditionBuildReady,
		CreateFailedReason: ReasonImageStreamCreationFailed,
	})

	// Create BuildConfig
	if isBuilderImageImportStrategy {
		if isBuildOptionBuildConfig {
			r.Resources = append(r.Resources, ConsoleAppResource{
				Type:               BuildConfig,
				Object:             newBuildConfig(r.consoleApplication),
				ConditionType:      ConditionBuildReady,
				CreateFailedReason: ReasonBuildConfigCreationFailed,
			})
		}
	}

	// Create Workload
	if isDeploymentResourceType {
		r.Resources = append(r.Resources, ConsoleAppResource{
			Type:               Deployment,
			Object:             newK8sDeployment(r.consoleApplication),
			ConditionType:      ConditionWorkloadReady,
			CreateFailedReason: ReasonWorkloadCreationFailed,
			NotReadyReason:     ReasonWorkloadNotReady,
			ReadyReason:        ReasonWorkloadReady,
		})
	}

	// Create Service
	r.Resources = append(r.Resources, ConsoleAppResource{
		Type:               Service,
		Object:             newService(r.consoleApplication),
		ConditionType:      ConditionServiceReady,
		CreateFailedReason: ReasonServiceCreationFailed,
		NotReadyReason:     ReasonServiceNotReady,
		ReadyReason:        ReasonServiceReady,
	})

	// Create Route
	if isRouteCreationEnabled {
		r.Resources = append(r.Resources, ConsoleAppResource{
			Type:               Route,
			Object:             newRoute(r.consoleApplication),
			ConditionType:      ConditionRouteReady,
			CreateFailedReason: ReasonRouteCreationFailed,
			NotReadyReason:     ReasonRouteNotReady,
			ReadyReason:        ReasonRouteReady,
		})
	}

	return r.Resources, nil
}

func (r *ResourceMapper) sanityCheck() error {
	if r.consoleApplication.Spec.ImportStrategy == appsv1alpha1.ImportStrategyBuilderImage {
		if r.consoleApplication.Spec.BuildConfiguration.BuilderImage.Image == "" ||
			r.consoleApplication.Spec.BuildConfiguration.BuilderImage.Tag == "" {
			return ErrBuilderImgNotProvided
		}
	}
	return nil
}

func (r *ResourceMapper) checkBuilderImageTagExists(imgStream *imagev1.ImageStream, tag string) (bool, error) {
	for _, tagEvent := range imgStream.Status.Tags {
		if tagEvent.Tag == tag {
			return true, nil
		}
	}
	return false, fmt.Errorf("ImageStreamTag %s:%s not found", imgStream.Name, tag)
}

func (r *ResourceMapper) CheckBuildStatus(builds *buildv1.BuildList) (requeueNeeded bool, status metav1.ConditionStatus, reason ConditionReason, message string) {
	if len(builds.Items) == 0 {
		return true, metav1.ConditionUnknown, ReasonBuildsNotFound, "Builds not created yet"
	}

	// Sort the builds by creation timestamp so the latest build will be the first element
	sort.Slice(builds.Items, func(i, j int) bool {
		return builds.Items[i].CreationTimestamp.After(builds.Items[j].CreationTimestamp.Time)
	})

	// Fetch the latest build
	latestBuild := builds.Items[0]
	switch latestBuild.Status.Phase {
	case buildv1.BuildPhaseNew, buildv1.BuildPhasePending, buildv1.BuildPhaseRunning:
		return true, metav1.ConditionUnknown, ConditionReason(latestBuild.Status.Phase), fmt.Sprintf("%s in progress", latestBuild.Name)
	case buildv1.BuildPhaseComplete:
		return false, metav1.ConditionTrue, ConditionReason(latestBuild.Status.Phase), fmt.Sprintf("%s has completed", latestBuild.Name)
	default:
		return false, metav1.ConditionFalse, ReasonBuildsFailed, fmt.Sprintf("Build Phase: %s", latestBuild.Status.Phase)
	}
}

func (r *ResourceMapper) CheckDeploymentStatus(deployment *appsv1.Deployment) (requeueNeeded bool, status metav1.ConditionStatus, reason ConditionReason, message string) {
	if deployment.Status.Conditions == nil {
		return true, metav1.ConditionUnknown, ReasonWorkloadNotReady, "Deployment not created yet"
	}

	availableStatusCondition := appsv1.DeploymentCondition{}
	progressingStatusCondition := appsv1.DeploymentCondition{}
	for _, condition := range deployment.Status.Conditions {
		switch condition.Type {
		case appsv1.DeploymentAvailable:
			availableStatusCondition = condition
		case appsv1.DeploymentProgressing:
			progressingStatusCondition = condition
		}
	}

	if progressingStatusCondition.Status == corev1.ConditionTrue {
		if availableStatusCondition.Status == corev1.ConditionFalse {
			return true, metav1.ConditionFalse, ReasonWorkloadNotReady, "Deployment Not Ready"
		} else if availableStatusCondition.Status == corev1.ConditionTrue {
			// Deployment is ready
			return false, metav1.ConditionTrue, ReasonWorkloadReady, "Deployment Ready"
		}
	}
	return false, metav1.ConditionFalse, ReasonWorkloadCreationFailed, "Deployment Progressing Failed"
}

func (r *ResourceMapper) GetResourceObject(resourceType ResourceType) client.Object {
	switch resourceType {
	case ImageStream:
		return &imagev1.ImageStream{}
	case BuildConfig:
		return &buildv1.BuildConfig{}
	case Deployment:
		return &appsv1.Deployment{}
	case Service:
		return &corev1.Service{}
	case Route:
		return &routev1.Route{}
	default:
		return nil
	}
}
