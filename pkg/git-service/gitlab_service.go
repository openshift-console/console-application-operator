package gitservice

import (
	"github.com/xanzy/go-gitlab"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isGLRepoReachable(g *GitService) {

	if g.secretValue == "" {
		g.logger.Error(nil, "Secret value not provided")
		g.status, g.reason = metav1.ConditionFalse, ReasonAccessTokenRequired
		return
	}

	client, err := gitlab.NewClient(g.secretValue)
	if err != nil {
		g.logger.Error(err, "Failed to create Gitlab client")
		g.status, g.reason = metav1.ConditionFalse, ReasonRepoNotReachable
		return
	}

	_, res, err := client.Branches.GetBranch(g.owner+"/"+g.repo, g.reference)
	if err != nil {
		g.logger.Error(err, "Unsuccessful response from Gitlab API")
		g.status = metav1.ConditionFalse
		switch res.StatusCode {
		case 429:
			g.reason = ReasonRateLimitExceeded
		case 404:
			g.reason = ReasonRepoNotFound
		default:
			g.reason = ReasonRepoNotReachable
		}
		return
	}
	g.logger.Info("Successfully reached Gitlab API!")
	g.status, g.reason = metav1.ConditionTrue, ReasonSucceeded
}
