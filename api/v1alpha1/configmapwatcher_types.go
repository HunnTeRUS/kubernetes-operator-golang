/*
Copyright 2025.

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

// ConfigMapWatcherSpec defines the desired state of ConfigMapWatcher.
type ConfigMapWatcherSpec struct {
	// ConfigMapName is the name of the ConfigMap to watch
	ConfigMapName string `json:"configMapName"`

	// ConfigMapNamespace is the namespace where the ConfigMap is located
	ConfigMapNamespace string `json:"configMapNamespace"`

	// EventEndpoint is the URL where events will be sent when the ConfigMap changes
	EventEndpoint string `json:"eventEndpoint"`

	// EventSecretName is the name of the secret containing credentials for the event endpoint
	// +optional
	EventSecretName string `json:"eventSecretName,omitempty"`

	// EventSecretNamespace is the namespace where the event secret is located
	// +optional
	EventSecretNamespace string `json:"eventSecretNamespace,omitempty"`
}

// ConfigMapWatcherStatus defines the observed state of ConfigMapWatcher.
type ConfigMapWatcherStatus struct {
	// LastConfigMapVersion is the last observed version of the ConfigMap
	LastConfigMapVersion string `json:"lastConfigMapVersion,omitempty"`

	// LastEventSent is the timestamp of the last event sent
	LastEventSent metav1.Time `json:"lastEventSent,omitempty"`

	// Conditions represent the latest available observations of the ConfigMapWatcher's current state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConfigMapWatcher is the Schema for the configmapwatchers API.
type ConfigMapWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigMapWatcherSpec   `json:"spec,omitempty"`
	Status ConfigMapWatcherStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigMapWatcherList contains a list of ConfigMapWatcher.
type ConfigMapWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMapWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigMapWatcher{}, &ConfigMapWatcherList{})
}
