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
	// ConditionBuildReady is True if the build config is created
	ConditionBuildReady ConditionType = "BuildReady"
	// ConditionWorkloadReady is True if the workload is ready
	ConditionWorkloadReady ConditionType = "WorkloadReady"
	// ConditionServiceReady is True if the service is ready
	ConditionServiceReady ConditionType = "ServiceReady"
	// ConditionRouteReady is True if the route is ready
	ConditionRouteReady ConditionType = "RouteReady"
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

	// ReasonImageStreamNotFound indicates the image stream is not found
	ReasonImageStreamNotFound ConditionReason = "ImageStreamNotFound"
	// ReasonReasonImageStreamCreationFailed indicates the image stream creation failed
	ReasonImageStreamCreationFailed ConditionReason = "ImageStreamCreationFailed"

	// ReasonBuildConfigCreated indicates the build config is created
	ReasonBuildConfigCreated ConditionReason = "BuildConfigCreated"
	// ReasonReasonBuildConfigCreationFailed indicates the build config creation failed
	ReasonBuildConfigCreationFailed ConditionReason = "BuildConfigCreationFailed"
	// ReasonReasonBuildsNotFound indicates the builds are not found
	ReasonBuildsNotFound ConditionReason = "BuildsNotFound"
	// ReasonBuildsFailed indicates the builds failed
	ReasonBuildsFailed ConditionReason = "BuildsFailed"

	// ReasonWorkloadCreationFailed indicates the deployment creation failed
	ReasonWorkloadCreationFailed ConditionReason = "WorkloadCreationFailed"
	// ReasonWorkloadNotFound indicates the deployment is not found
	ReasonWorkloadNotFound ConditionReason = "WorkloadNotFound"
	// ReasonWorkloadNotReady indicates the deployment is not ready
	ReasonWorkloadNotReady ConditionReason = "WorkloadNotReady"
	// ReasonWorkloadReady indicates the deployment is ready
	ReasonWorkloadReady ConditionReason = "WorkloadReady"

	// ReasonServiceCreationFailed indicates the service creation failed
	ReasonServiceCreationFailed ConditionReason = "ServiceCreationFailed"
	// ReasonServiceNotFound indicates the service is not found
	ReasonServiceNotFound ConditionReason = "ServiceNotFound"
	// ReasonServiceNotReady indicates the service is not ready
	ReasonServiceNotReady ConditionReason = "ServiceNotReady"
	// ReasonServiceReady indicates the service is ready
	ReasonServiceReady ConditionReason = "ServiceReady"

	// ReasonRouteCreationFailed indicates the route creation failed
	ReasonRouteCreationFailed ConditionReason = "RouteCreationFailed"
	// ReasonRouteNotFound indicates the route is not found
	ReasonRouteNotFound ConditionReason = "RouteNotFound"
	// ReasonRouteNotReady indicates the route is not ready
	ReasonRouteNotReady ConditionReason = "RouteNotReady"
	// ReasonRouteReady indicates the route is ready
	ReasonRouteReady ConditionReason = "RouteReady"

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
