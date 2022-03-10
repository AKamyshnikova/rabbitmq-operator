/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RabbitMQSpec defines the desired state of RabbitMQ
type RabbitMQSpec struct {
	Replicas int32  `json:"replicas"`
	Image    string `json:"image"`
	// +kubebuilder:default:="Always"
	ImagePullPolicy  corev1.PullPolicy            `json:"imagePullPolicy,omitempty"`
	ServiceAccount   string                       `json:"service_account"`
	DiscoveryService string                       `json:"discovery_service,omitempty"`
	Vhost            string                       `json:"vhost,omitempty"`
	DataVolumeSize   resource.Quantity            `json:"data_volume_size"`
	DataStorageClass string                       `json:"dataStorageClass,omitempty"`
	Affinity         *corev1.Affinity             `json:"affinity,omitempty"`
	Resources        *corev1.ResourceRequirements `json:"resources,omitempty"`
	DefaultUsername  string                       `json:"defaultUsername,omitempty"`
	DefaultPassword  string                       `json:"defaultPassword,omitempty"`
	DefaultVHost     string                       `json:"defaultVHost,omitempty"`
	ExporterImage    string                       `json:"exporter_image,omitempty"`
	ExporterPort     int32                        `json:"exporter_port,omitempty"`
}

// RabbitMQStatus defines the observed state of RabbitMQ
type RabbitMQStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RabbitMQ is the Schema for the rabbitmqs API
type RabbitMQ struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSpec   `json:"spec,omitempty"`
	Status RabbitMQStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RabbitMQList contains a list of RabbitMQ
type RabbitMQList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQ `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQ{}, &RabbitMQList{})
}
