package gitservice

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isGHRepoReachable(g *GitService) {
	ctx := context.Background()

	oauth_client := (*http.Client)(nil)
	if g.secretValue != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.secretValue},
		)
		oauth_client = oauth2.NewClient(ctx, ts)
	}

	client := github.NewClient(oauth_client)

	_, resp, err := client.Repositories.GetBranch(ctx, g.owner, g.repo, g.reference)
	if err != nil {
		g.logger.Error(err, "Unsuccessful response from Github API")
		g.status = "False"
		switch resp.StatusCode {
		case 403:
			g.reason = ReasonRateLimitExceeded
		case 404:
			g.reason = ReasonRepoNotFound
		default:
			g.reason = ReasonRepoNotReachable
		}
		return
	}
	g.logger.Info("Successfully reached Github API!")
	g.status, g.reason = metav1.ConditionTrue, ReasonSucceeded
}
