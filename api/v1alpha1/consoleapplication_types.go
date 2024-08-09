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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Git struct {
	Url             string `json:"url,omitempty"`
	ContextDir      string `json:"contextDir,omitempty"`
	Reference       string `json:"reference,omitempty"`
	SourceSecretRef string `json:"sourceSecretRef,omitempty"`
}

type BuildConfiguration struct {
	BuilderImage BuilderImage `json:"builderImage,omitempty"`
	BuildOption  string       `json:"buildOption,omitempty"`
	Env          []Env        `json:"env,omitempty"`
}

type BuilderImage struct {
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}

type Env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Expose struct {
	TargetPort  *int32 `json:"targetPort,omitempty"`
	CreateRoute bool   `json:"createRoute,omitempty"`
}

type DeploymentConfiguration struct {
	ResourceType string `json:"resourceType,omitempty"`
	Env          []Env  `json:"env,omitempty"`
	Expose       Expose `json:"expose,omitempty"`
}

// ConsoleApplicationSpec defines the desired state of ConsoleApplication
type ConsoleApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ConsoleApplication. Edit consoleapplication_types.go to remove/update
	ApplicationName         string                  `json:"applicationName,omitempty"`
	Git                     Git                     `json:"git,omitempty"`
	ImportStrategy          string                  `json:"importStrategy,omitempty"`
	BuildConfiguration      BuildConfiguration      `json:"buildConfiguration,omitempty"`
	DeploymentConfiguration DeploymentConfiguration `json:"deploymentConfiguration,omitempty"`
}

// ConsoleApplicationStatus defines the observed state of ConsoleApplication
type ConsoleApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
