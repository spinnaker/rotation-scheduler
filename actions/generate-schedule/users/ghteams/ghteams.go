package ghteams

import (
	"context"
	"net/http"

	"github.com/google/go-github/v30/github"
	"github.com/spinnaker/rotation-scheduler/actions/generate-schedule/users"
)

type GitHubTeamsUserSource = users.StaticSource

// NewGitHubTeamsUserSource fetches the current GitHub usernames from the specified team. The http.Client implementation
// must attach a GitHub personal access token (with the 'read:org' scope) to the request, such as one from the oauth2
// package.
func NewGitHubTeamsUserSource(client *http.Client, orgName, teamName string) (*GitHubTeamsUserSource, error) {
	ghClient := github.NewClient(client)

	ghUsers, _, err := ghClient.Teams.ListTeamMembersBySlug(context.Background(), orgName, teamName, nil /*opts*/)
	if err != nil {
		return nil, err
	}

	var logins []string
	for _, ghu := range ghUsers {
		logins = append(logins, ghu.GetLogin())
	}

	return users.NewStaticSource(logins...), nil
}
