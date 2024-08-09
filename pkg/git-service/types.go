package gitservice

type (

	// GitProvider is the type of Git provider
	GitProvider string

	// GitConditionReason is a static/programmatic representation of the cause of a status condition.
	GitConditionReason string
)

const (
	// Github is the Github provider
	Github GitProvider = "github"

	// Gitlab is the Gitlab provider
	Gitlab GitProvider = "gitlab"

	// Unknown is the unknown provider
	Unknown GitProvider = "unknown"

	// ReasonProcessing indicates the condition is processing
	ReasonProcessing GitConditionReason = "Processing"

	// ReasonSucceeded indicates the condition is succeeded
	ReasonSucceeded GitConditionReason = "Succeeded"

	// ReasonRepoNotFound indicates the repository was not found
	ReasonRepoNotFound GitConditionReason = "RepoNotFound"

	// ReasonRepoNotReachable indicates the repository is not reachable
	ReasonRepoNotReachable GitConditionReason = "RepoNotReachable"

	// ReasonRateLimitExceeded indicates the rate limit for fetching Github repository has been exceeded
	ReasonRateLimitExceeded GitConditionReason = "RateLimitExceeded"

	// ReasonUnsupportedGitType indicates the Git type is not supported
	ReasonUnsupportedGitType GitConditionReason = "UnsupportedGitType"

	// ReasonInvalidGitURL indicates the Git URL is invalid
	ReasonInvalidGitURL GitConditionReason = "InvalidGitURL"

	// ReasonAccessTokenRequired indicates the Gitlab URL is not reachable because it requires an access token
	ReasonAccessTokenRequired GitConditionReason = "AccessTokenRequired"
)

// String casts the value to string.
// "r.String()" and "string(r)" are equivalent.
func (r GitConditionReason) String() string {
	return string(r)
}
