package v1alpha1


import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// SecurityChaos is the Schema for the SecurityChaos API
type SecurityChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityChaosSpec   `json:"spec"`
	Status SecurityChaosStatus `json:"status,omitempty"`
}

// SecurityChaosSpec is the content of the specification for a SecurityChaos
type SecurityChaosSpec struct {
	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`
}

// SecurityChaosStatus represents the status of a SecurityChaos
type SecurityChaosStatus struct {
	ChaosStatus `json:",inline"`
}
