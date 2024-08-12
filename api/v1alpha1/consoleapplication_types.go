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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImportStrategyType defines the type of import strategy
// +kubebuilder:validation:Enum=BuilderImage;Dockerfile
type ImportStrategyType string

const (
	// ImportStrategyBuilderImage is used for building applications using a builder image
	ImportStrategyBuilderImage ImportStrategyType = "BuilderImage"

	// ImportStrategyDockerfile is used for building applications using a Dockerfile
	ImportStrategyDockerfile ImportStrategyType = "Dockerfile"
)

// BuildOption defines the type of build option
// +kubebuilder:validation:Enum=BuildConfig
type BuildOption string

const (
	// BuildOptionBuildConfig is used for building applications using OpenShift BuildConfig(v1) Resource
	BuildOptionBuildConfig BuildOption = "BuildConfig"
)

// ResourceType defines the type of resource to deploy the application
// +kubebuilder:validation:Enum=Deployment
type ResourceType string

const (
	// ResourceTypeDeployment is used for deploying applications as a Kubernetes Deployment
	ResourceTypeDeployment ResourceType = "Deployment"
)

// Git defines the git repository details
type Git struct {
	// URL is the git repository URL
	// +kubebuilder:validation:Required
	Url string `json:"url,omitempty"`

	// ContextDir is the directory within the git repository to use as the context for the build
	// +kubebuilder:default=/
	// +kubebuilder:validation:Optional
	ContextDir string `json:"contextDir,omitempty"`

	// Reference is the branch, tag, or commit
	// +kubebuilder:default=main
	// +kubebuilder:validation:Optional
	Reference string `json:"reference,omitempty"`

	// SourceSecretRef is the reference to the secret containing the git credentials
	// +kubebuilder:validation:Optional
	SourceSecretRef string `json:"sourceSecretRef,omitempty"`
}

// BuildConfiguration defines the build configuration
type BuildConfiguration struct {
	// BuilderImage is the builder image to use for building the application
	// +kubebuilder:validation:Optional
	BuilderImage BuilderImage `json:"builderImage,omitempty"`

	// BuildOption is the build option to use for building the application
	// +kubebuilder:default=BuildConfig
	// +kubebuilder:validation:Optional
	BuildOption BuildOption `json:"buildOption,omitempty"`

	// Env is the environment variables to set during the build
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// BuilderImage defines the builder image details to use for building the application when ImportStrategyType is "BuilderImage"
type BuilderImage struct {
	// Image is the builder image to use for building the application
	// +kubebuilder:validation:Required
	Image string `json:"image,omitempty"`

	// Tag is the tag of the builder image to use for building the application
	// +kubebuilder:validation:Required
	Tag string `json:"tag,omitempty"`
}

// Expose defines the OpenShift Route configuration
type Expose struct {
	// TargetPort is the port where the application is running
	// +kubebuilder:validation:Required
	// +kubebuilder:default=8080
	TargetPort *int32 `json:"targetPort,omitempty"`

	// CreateRoute is a flag to create a route for the application
	// +kubebuilder:default=true
	// +kubebuilder:validation:Optional
	CreateRoute bool `json:"createRoute,omitempty"`
}

// DeploymentConfiguration defines the deployment configuration
type DeploymentConfiguration struct {
	// ResourceType is the type of resource to deploy the application
	// +kubebuilder:validation:Required
	ResourceType ResourceType `json:"resourceType,omitempty"`

	// Env is the environment variables to set during the deployment
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Expose is the OpenShift Route configuration
	// +kubebuilder:validation:Required
	Expose Expose `json:"expose,omitempty"`
}

// ConsoleApplicationSpec defines the desired state of ConsoleApplication
type ConsoleApplicationSpec struct {
	// ApplicationName is the name of the application. This is used to group the resources created for the application.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="console-application"
	ApplicationName string `json:"applicationName,omitempty"`

	// Git is the git repository details
	// +kubebuilder:validation:Required
	Git Git `json:"git,omitempty"`

	// ImportStrategy is the import strategy to use for importing the application
	// +kubebuilder:validation:Required
	ImportStrategy ImportStrategyType `json:"importStrategy,omitempty"`

	// BuildConfiguration is the build configuration
	// +kubebuilder:validation:Required
	BuildConfiguration BuildConfiguration `json:"buildConfiguration,omitempty"`

	// DeploymentConfiguration is the deployment configuration
	// +kubebuilder:validation:Required
	DeploymentConfiguration DeploymentConfiguration `json:"deploymentConfiguration,omitempty"`
}

// ConsoleApplicationStatus defines the observed state of ConsoleApplication
type ConsoleApplicationStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConsoleApplication is the Schema for the consoleapplications API
type ConsoleApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsoleApplicationSpec   `json:"spec,omitempty"`
	Status ConsoleApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConsoleApplicationList contains a list of ConsoleApplication
type ConsoleApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsoleApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConsoleApplication{}, &ConsoleApplicationList{})
}
