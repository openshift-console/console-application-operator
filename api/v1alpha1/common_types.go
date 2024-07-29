package v1alpha1

type (
	// ConditionType defines what condition type this is.
	ConditionType string

	// ConditionReason is a static/programmatic representation of the cause of a status condition.
	ConditionReason string
)

const (

	// ConditionProgressing is True when the operator is progressing
	ConditionProgressing ConditionType = "Progressing"

	// ConditionApplicationReady is True when the application is ready
	ConditionApplicationReady ConditionType = "ApplicationReady"

	// ConditionGitRepoReachable is True if the Git repository is reachable
	ConditionGitRepoReachable ConditionType = "GitRepoReachable"

	// ConditionOperatorDegraded is True if the operator is in a degraded state
	ConditionOperatorDegraded ConditionType = "OperatorDegraded"

	// ConditionResourceNotFound is True if the resource is not found
	ConditionResourceNotFound ConditionType = "ResourceNotFound"

	// ReasonOperatorResourceNotAvailable indicates the operator resource is not available
	ReasonOperatorResourceNotAvailable ConditionReason = "OperatorResourceNotAvailable"

	// ReasonSecretResourceNotFound indicates the secret resource is not found
	ReasonSecretResourceNotFound ConditionReason = "SecretResourceNotFound"

	// ReasonAllResourcesReady indicates all resources are ready
	ReasonAllResourcesReady ConditionReason = "AllResourcesReady"

	// ReasonReconcileFailed indicates the reconciliation failed
	ReasonReconcileFailed ConditionReason = "ReconcileFailed"

	// ReasonReconcileCompleted indicates the reconciliation completed
	ReasonReconcileCompleted ConditionReason = "ReconcileCompleted"
)

// String casts the value to string.
// "c.String()" and "string(c)" are equivalent.
func (c ConditionType) String() string {
	return string(c)
}

// String casts the value to string.
// "r.String()" and "string(r)" are equivalent.
func (r ConditionReason) String() string {
	return string(r)
}
