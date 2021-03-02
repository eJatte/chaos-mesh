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

// SecurityChaosAction represents the chaos action about security.
type SecurityChaosAction string

const (
	// RunAsRootAction represents the chaos action of creating a container as root.
	RunAsRootAction SecurityChaosAction = "run-as-root"
	// RunAsPrivilegedAction represents the chaos action of creating a privileged container.
	RunAsPrivilegedAction SecurityChaosAction = "run-as-privileged"
	// TestAction represents the chaos action of test
	TestAction SecurityChaosAction = "test"
)

// SecurityChaosMessage
type SecurityChaosMessage string

const (
	AttackSucceededMessage SecurityChaosMessage = "ATTACK_SUCCEEDED"
	AttackFailedMessage SecurityChaosMessage = "ATTACK_FAILED"
)

// SecurityChaosSpec is the content of the specification for a SecurityChaos
type SecurityChaosSpec struct {
	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the specific security chaos action.
	// Supported action: run-as-root / run-as-privileged / test
	// Default action: run-as-root
	// +kubebuilder:validation:Enum=run-as-root;run-as-privileged;test
	Action SecurityChaosAction `json:"action"`

	// NameSpace defines the namespace that the chaos experiment should be applied in.
	// Default namespace: default
	// +optional
	NameSpace string `json:"namespace"`
}

// SecurityChaosStatus represents the status of a SecurityChaos
type SecurityChaosStatus struct {
	ChaosStatus `json:",inline"`
}
