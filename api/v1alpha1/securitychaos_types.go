package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
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
	// CreatePodAction represents the chaos action of creating a pod.
	CreatePodAction SecurityChaosAction = "create-pod"
	// DeleteFileAction represents the chaos action of attempting to delete a file that should not be deletable.
	DeleteFileAction SecurityChaosAction = "delete-file"
	// ListSecretsAction represents the chaos action of attempting to list secrets
	ListSecretsAction SecurityChaosAction = "list-secrets"
	// KubeletAPIAction represents the chaos action of attempting to access the kubelet api
	KubeletAPIAction SecurityChaosAction = "kubelet-api"
	// TestAction represents the chaos action of test
	TestAction SecurityChaosAction = "test"
)

// SecurityChaosMessage
type SecurityChaosMessage string

const (
	AttackSucceededMessage SecurityChaosMessage = "ATTACK_SUCCEEDED"
	AttackFailedMessage    SecurityChaosMessage = "ATTACK_FAILED"
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
	// Supported action: run-as-root / run-as-privileged / delete-file / list-secrets / create-pod / kubelet-api / test
	// Default action: run-as-root
	// +kubebuilder:validation:Enum=run-as-root;run-as-privileged;delete-file;list-secrets;create-pod;kubelet-api;test;
	Action SecurityChaosAction `json:"action"`

	// NameSpace defines the namespace that the chaos experiment should be applied in.
	// Default namespace: default
	// +optional
	NameSpace string `json:"namespace"`

	// Selector is used to select pods that are used to inject chaos action.
	// +optional
	Selector SelectorSpec `json:"selector"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +optional
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// IF `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// PvClaim specifies the persistent volume claim.
	// +optional
	PvClaim string `json:"pvclaim"`

	// UID specifies the uid to use in the experiment, needed in delete file experiment.
	// +optional
	UID int64 `json:"uid"`

	// GID specifies the gid to use in the experiment, needed in delete file experiment.
	// +optional
	GID int64 `json:"gid"`

	// User specifies a kubernetes user. Used in the list secrets experiment.
	// +optional
	User string `json:"user"`

	// PodSecurityContext specifies a pod security contex, used in the create-pod experiment
	// +optional
	PodSecurityContext v1.PodSecurityContext `json:"podsecuritycontext"`

	// SecurityContext specifies a security contex, used in the create-pod experiment
	// +optional
	SecurityContext v1.SecurityContext `json:"securitycontext"`

	// Node specifies the name of a kubernetes node, used in the kubelet api experiment.
	// +optional
	Node string `json:"node"`
}

// SecurityChaosStatus represents the status of a SecurityChaos
type SecurityChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *SecurityChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

func (in *SecurityChaosSpec) GetMode() PodMode {
	return in.Mode
}

func (in *SecurityChaosSpec) GetValue() string {
	return in.Value
}
