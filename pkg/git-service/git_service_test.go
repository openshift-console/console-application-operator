package gitservice

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifyGitType(t *testing.T) {
	tests := []struct {
		name   string
		gitURL string
		want   GitProvider
	}{
		{"Is Github Repo", "http://www.github.com/hello/world", Github},
		{"Is Gitlab Repo", "http://www.gitlab.com/hello/world", Gitlab},
		{"Is Unknown Repo", "http://www.example.com/hello/world", Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := identifyGitType(tt.gitURL)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetOwnerAndRepo(t *testing.T) {
	tests := []struct {
		name   string
		gitURL string
		owner  string
		repo   string
		err    error
	}{
		{"With https and www", "https://www.github.com/hello/world", "hello", "world", nil},
		{"Without http or https or www", "github.com/hello/world", "hello", "world", nil},
		{"URL without repoName", "https://www.github.com/hello", "", "", errors.New("InvalidGitURL")},
		{"Gitlab Repo", "http://www.gitlab.com/hello/world", "hello", "world", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, repo, err := getOwnerAndRepo(tt.gitURL)
			if tt.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.owner, username)
				assert.Equal(t, tt.repo, repo)
			}
		})
	}
}
