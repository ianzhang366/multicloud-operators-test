package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppTestSpec defines the desired state of AppTest
type AppTestSpec struct {
	Resources []AppTestResources `json:"resources"`
}

// AppTestStatus defines the observed state of AppTest
type AppTestStatus struct {
	TestStatus      string             `json:"testStatus,omitempty"`
	FailedResources []AppTestResources `json:"failedResources,omitempty"`
}

// AppTestResources defines the resources that need to be checked
type AppTestResources struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Cluster defaults to hub cluster if not specified
	Cluster string `json:"cluster"`

	// String to interface mapping to define a flexible spec profile
	DesiredStatus map[string]interface{} `json:"desiredStatus"`

	// Updated by the controller to show current status of the defined resource
	CurrentStatus map[string]interface{} `json:"currentStatus,omitempty"`
	Messages      []string               `json:"messages,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppTest is the Schema for the apptests API
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.testStatus`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apptests,scope=Namespaced
type AppTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppTestSpec   `json:"spec,omitempty"`
	Status AppTestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppTestList contains a list of AppTest
type AppTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppTest{}, &AppTestList{})
}
