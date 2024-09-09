package v1alpha1

type (
	// ConditionType defines what condition type this is.
	ConditionType string
	// ConditionReason is a static/programmatic representation of the cause of a status condition.
	ConditionReason string
)

/* ConditionType is used to define the type of status of the operator. */
const (
	// ConditionOperatorDegraded is True if the operator is in a degraded state
	ConditionOperatorDegraded ConditionType = "OperatorDegraded"
	// ConditionProgressing is True when the operator is progressing
	ConditionProgressing ConditionType = "Progressing"
	// ConditionGitRepoReachable is True if the Git repository is reachable
	ConditionGitRepoReachable ConditionType = "GitRepoReachable"
	// ConditionReady is True when the application is ready
	ConditionReady ConditionType = "Ready"
)

/* ConditionReason is used to define the reason for the status of the operator. */
const (
	// ReasonOperatorResourceNotAvailable indicates the operator resource is not available
	ReasonOperatorResourceNotAvailable ConditionReason = "OperatorResourceNotAvailable"

	// ReasonSecretResourceNotFound indicates the secret resource is not found
	ReasonSecretResourceNotFound ConditionReason = "SecretResourceNotFound"

	// ReasonInit indicates the resource is initializing
	ReasonInit ConditionReason = "Init"

	// ReasonRequirementsNotMet indicates the reconciliation failed
	ReasonRequirementsNotMet ConditionReason = "RequirementsNotMet"
	//ReasonRequirementsBeingMet indicates the reconciliation is in progress
	ReasonRequirementsBeingMet ConditionReason = "RequirementsBeingMet"
	// ReasonRequirementsMet indicates the reconciliation completed
	ReasonRequirementsMet ConditionReason = "RequirementsMet"

	// ReasonAllResourcesReady indicates all resources are ready
	ReasonAllResourcesReady ConditionReason = "AllResourcesReady"
)

// String casts the value to string.
func (c ConditionType) String() string {
	return string(c)
}

// String casts the value to string.
func (r ConditionReason) String() string {
	return string(r)
}
