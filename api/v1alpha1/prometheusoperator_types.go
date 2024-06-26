package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrometheusOperatorSpec defines the desired state of PrometheusOperator
type PrometheusOperatorSpec struct {
	Size        int32  `json:"size"`
	Image       string `json:"image"`
	StorageSize string `json:"storageSize"`
}

// PrometheusOperatorStatus defines the observed state of PrometheusOperator
type PrometheusOperatorStatus struct {
	Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PrometheusOperator is the Schema for the prometheusoperators API
type PrometheusOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrometheusOperatorSpec   `json:"spec,omitempty"`
	Status PrometheusOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PrometheusOperatorList contains a list of PrometheusOperator
type PrometheusOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrometheusOperator{}, &PrometheusOperatorList{})
}
