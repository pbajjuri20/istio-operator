package v1alpha3

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// StatusType represents the status for a control plane, component, or resource
type StatusType struct {
	ObservedGeneration int64       `json:"observedGeneration,omitempty"`
	Conditions         []Condition `json:"conditions,omitempty"`
}

// NewStatus returns a new StatusType object
func NewStatus() StatusType {
	return StatusType{Conditions: make([]Condition, 0, 3)}
}

// ComponentStatus represents the status for a component
type ComponentStatus struct {
	StatusType     `json:",inline"`
	ResourceStatus map[ResourceKey]*StatusType `json:"resourceStatus,omitempty"`
}

// NewComponentStatus returns a new ComponentStatus object
func NewComponentStatus() *ComponentStatus {
	return &ComponentStatus{StatusType: NewStatus(), ResourceStatus: map[ResourceKey]*StatusType{}}
}

// ConditionType represents the type of the condition.  Condition stages are:
// Initialized, Installed, Reconciled
type ConditionType string

const (
	// ConditionTypeInitialized signifies the whether or not the controller has
	// initialized the CR.
	ConditionTypeInitialized ConditionType = "Initialized"
	// ConditionTypeInstalled signifies the whether or not the controller has
	// installed the resources defined through the CR.
	ConditionTypeInstalled ConditionType = "Installed"
	// ConditionTypeReconciled signifies the whether or not the controller has
	// reconciled the resources defined through the CR.
	ConditionTypeReconciled ConditionType = "Reconciled"
)

// ConditionStatus represents the status of the condition
type ConditionStatus string

const (
	// ConditionStatusTrue represents completion of the condition, e.g.
	// Initialized=True signifies that initialization has occurred.
	ConditionStatusTrue ConditionStatus = "True"
	// ConditionStatusFalse represents incomplete status of the condition, e.g.
	// Initialized=False signifies that initialization has not occurred or has
	// failed.
	ConditionStatusFalse ConditionStatus = "False"
	// ConditionStatusUnknown represents unknown completion of the condition, e.g.
	// Initialized=Unknown signifies that initialization may or may not have been
	// completed.
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

// ConditionReason represents a short message indicating how the condition came
// to be in its present state.
type ConditionReason string

const (
	// ConditionReasonDeletionSuccessful ...
	ConditionReasonDeletionSuccessful ConditionReason = "DeletionSuccessful"
	// ConditionReasonDeletionError ...
	ConditionReasonDeletionError ConditionReason = "DeletionError"
	// ConditionReasonInstallSuccessful ...
	ConditionReasonInstallSuccessful ConditionReason = "InstallSuccessful"
	// ConditionReasonInstallError ...
	ConditionReasonInstallError ConditionReason = "InstallError"
	// ConditionReasonReconcileSuccessful ...
	ConditionReasonReconcileSuccessful ConditionReason = "ReconcileSuccessful"
	// ConditionReasonReconcileError ...
	ConditionReasonReconcileError ConditionReason = "ReconcileError"
	// ConditionReasonUninstallSuccessful ...
	ConditionReasonUninstallSuccessful ConditionReason = "UninstallSuccessful"
	// ConditionReasonUninstallError ...
	ConditionReasonUninstallError ConditionReason = "UninstallError"
	// ConditionReasonUpdateSuccessful ...
	ConditionReasonUpdateSuccessful ConditionReason = "UpdateSuccessful"
	// ConditionReasonUpdateError ...
	ConditionReasonUpdateError ConditionReason = "UpdateError"
)

// Condition represents a specific condition on a resource
type Condition struct {
	Type               ConditionType   `json:"type,omitempty"`
	Status             ConditionStatus `json:"status,omitempty"`
	Reason             ConditionReason `json:"reason,omitempty"`
	Message            string          `json:"message,omitempty"`
	LastTransitionTime metav1.Time     `json:"lastTransitionTime,omitempty"`
}

// GetCondition removes a condition for the list of conditions
func (s *StatusType) GetCondition(conditionType ConditionType) Condition {
	if s == nil {
		return Condition{Type: conditionType, Status: ConditionStatusUnknown}
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			return s.Conditions[i]
		}
	}
	return Condition{Type: conditionType, Status: ConditionStatusUnknown}
}

// SetCondition sets a specific condition in the list of conditions
func (s *StatusType) SetCondition(condition Condition) *StatusType {
	if s == nil {
		return nil
	}
	now := metav1.Now()
	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			if s.Conditions[i].Status != condition.Status {
				condition.LastTransitionTime = now
			} else {
				condition.LastTransitionTime = s.Conditions[i].LastTransitionTime
			}
			s.Conditions[i] = condition
			return s
		}
	}

	// If the condition does not exist,
	// initialize the lastTransitionTime
	condition.LastTransitionTime = now
	s.Conditions = append(s.Conditions, condition)
	return s
}

// RemoveCondition removes a condition for the list of conditions
func (s *StatusType) RemoveCondition(conditionType ConditionType) *StatusType {
	if s == nil {
		return nil
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			s.Conditions = append(s.Conditions[:i], s.Conditions[i+1:]...)
			return s
		}
	}
	return s
}

// ResourceKey is a typedef for key used in ManagedGenerations.  It is a string
// with the format: namespace/name=group/version,kind
type ResourceKey string

// NewResourceKey for the object and type
func NewResourceKey(o metav1.Object, t metav1.Type) ResourceKey {
	return ResourceKey(fmt.Sprintf("%s/%s=%s,%s", o.GetNamespace(), o.GetName(), t.GetAPIVersion(), t.GetKind()))
}

// ToUnstructured returns a an Unstructured object initialized with Namespace,
// Name, APIVersion, and Kind fields from the ResourceKey
func (key ResourceKey) ToUnstructured() *unstructured.Unstructured {
	// ResourceKey is guaranteed to be at least "/=," meaning we are guaranteed
	// to get two elements in all of the splits
	retval := &unstructured.Unstructured{}
	parts := strings.SplitN(string(key), "=", 2)
	nn := strings.SplitN(parts[0], "/", 2)
	gvk := strings.SplitN(parts[1], ",", 2)
	retval.SetNamespace(nn[0])
	retval.SetName(nn[1])
	retval.SetAPIVersion(gvk[0])
	retval.SetKind(gvk[1])
	return retval
}

// FindResourcesOfKind returns all the specified kind.  Note, this does not account for group or version.
func (s *ComponentStatus) FindResourcesOfKind(kind string) map[ResourceKey]*StatusType {
	resources := map[ResourceKey]*StatusType{}
	suffix := ","+kind
	for key, status := range s.ResourceStatus {
		if strings.HasSuffix(string(key), suffix) {
			resources[key] = status
		}
	}
	return resources
}
