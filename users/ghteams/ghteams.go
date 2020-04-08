package ghteams

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-github/v30/github"
	"github.com/spinnaker/rotation-scheduler/users"
)

type GitHubTeamsUserSource = users.StaticSource

// NewGitHubTeamsUserSource fetches the current GitHub usernames from the specified team. The http.Client implementation
// must attach a GitHub personal access token (with the 'read:org' scope) to the request, such as one from the oauth2
// package.
func NewGitHubTeamsUserSource(client *http.Client, orgName, teamName string, emailDomains ...string) (*GitHubTeamsUserSource, error) {
	ghClient := github.NewClient(client)
	ctx := context.Background()

	ghUsers, _, err := ghClient.Teams.ListTeamMembersBySlug(ctx, orgName, teamName, nil /*opts*/)
	if err != nil {
		return nil, err
	}

	loginsAndEmails := make([]string, len(ghUsers))
	for i, ghu := range ghUsers {
		login := ghu.GetLogin()
		if len(emailDomains) > 0 {
			if userDetails, _, err := ghClient.Users.Get(ctx, login); err == nil { // Just use login if an error occurs.
				if len(emailDomains) == 1 && emailDomains[0] == "*" {
					if userDetails.GetEmail() != "" {
						login = userDetails.GetEmail()
					}
				} else {
					for _, d := range emailDomains {
						if strings.HasSuffix(userDetails.GetEmail(), d) {
							login = userDetails.GetEmail()
							break
						}
					}
				}
			}
		}
		loginsAndEmails[i] = login
	}

	return users.NewStaticSource(loginsAndEmails...), nil
}
