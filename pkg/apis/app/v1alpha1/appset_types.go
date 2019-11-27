package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppSetSpec defines the desired state of AppSet
// +k8s:openapi-gen=true
type AppSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Image string `json:"image"`
	Replicas int32 `json:"replicas"`
	ServiceType string `json:"servicetype"`
}

// AppSetStatus defines the observed state of AppSet
// +k8s:openapi-gen=true
type AppSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Replicas int32    `json:"replicas"`
	PodNames []string `json:"podNames"`

}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppSet is the Schema for the appsets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=appsets,scope=Namespaced
type AppSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSetSpec   `json:"spec,omitempty"`
	Status AppSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppSetList contains a list of AppSet
type AppSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppSet{}, &AppSetList{})
}
