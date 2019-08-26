package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RabbitMQSpec defines the desired state of RabbitMQ
// +k8s:openapi-gen=true
type RabbitMQSpec struct {
	Replicas 			int32 	`json:"replicas"`
	Image				string	`json:"image"`
	ServiceAccount		string	`json:"service_account"`
	DiscoveryService	string 	`json:"discovery_service"`
	Vhost 				string 	`json:"vhost,omitempty"`
	DataVolumeSize		resource.Quantity	`json:"data_volume_size"`
}

// RabbitMQStatus defines the observed state of RabbitMQ
// +k8s:openapi-gen=true
type RabbitMQStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitMQ is the Schema for the rabbitmqs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type RabbitMQ struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSpec   `json:"spec,omitempty"`
	Status RabbitMQStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitMQList contains a list of RabbitMQ
type RabbitMQList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQ `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQ{}, &RabbitMQList{})
}
