package resourcemapper

type (
	// ConditionType defines what condition type this is.
	ConditionType string
	// ConditionReason is a static/programmatic representation of the cause of a status condition.
	ConditionReason string

	// ErrorType is a static/programmatic representation of the cause of an error.
	ErrorType string
)

/* ConditionType is used to define the type of status of the operator. */
const (
	// ConditionBuildReady is True if the build config is created
	ConditionBuildReady ConditionType = "BuildReady"
	// ConditionWorkloadReady is True if the workload is ready
	ConditionWorkloadReady ConditionType = "WorkloadReady"
	// ConditionServiceReady is True if the service is ready
	ConditionServiceReady ConditionType = "ServiceReady"
	// ConditionRouteReady is True if the route is ready
	ConditionRouteReady ConditionType = "RouteReady"
)

/* ConditionReason is used to define the reason for the status of the operator. */
const (
	// ReasonRequiredResourcesNotFound indicates the required resources are not found
	ReasonRequiredResourcesNotFound ConditionReason = "RequiredResourcesNotFound"
	// ReasonRequiredResourceStatusCheckFailed indicates the required resource status check failed
	ReasonRequiredResourceStatusCheckFailed ConditionReason = "RequiredResourceStatusCheckFailed"

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
)

/* ErrorType is used to define the type of error. */
const (
	// ErrorTypeResourceNotFound indicates the resource is not found
	ErrCreateResource ErrorType = "Error creating resource: "

	ErrCreateImgStream ErrorType = "Error creating ImageStream: "

	ErrFetchImgStream ErrorType = "Error fetching ImageStream: "

	ErrBuilderImgNotProvided ErrorType = "Builder image and tag not provided"
)

// String casts the value to string.
func (c ConditionType) String() string {
	return string(c)
}

// Error implements the error interface for ErrorType.
func (e ErrorType) Error() string {
	return string(e)
}

// String casts the value to string.
func (e ErrorType) String() string {
	return string(e)
}

// String casts the value to string.
func (r ConditionReason) String() string {
	return string(r)
}
