package ghteams

import (
	"context"
	"reflect"
	"testing"

	"cloud.google.com/go/httpreplay"
)

var (
	wellKnownDomains = []string{
		"armory.io",
		"google.com",
		"netflix.com",
	}
)

// Replay generated from:
// go run rotation.go schedule generate --record ./users/ghteams/testing/teams.replay --github spinnaker,build-cops,$GITHUB_TOKEN --start 2020-05-01 --stop 2020-06-01 --domains armory.io,netflix.com,google.com
func TestNewGitHubTeamsUserSource(t *testing.T) {
	r, err := httpreplay.NewReplayer("testing/teams.replay")
	if err != nil {
		t.Fatalf("error creating replayer: %v", err)
	}

	client, err := r.Client(context.Background())
	if err != nil {
		t.Fatalf("error creating replayer client: %v", err)
	}

	ghUserSrc, err := NewGitHubTeamsUserSource(client, "spinnaker", "build-cops", wellKnownDomains...)
	if err != nil {
		t.Fatalf("error getting users: %v", err)
	}

	first := ghUserSrc.NextUser()
	got := []string{first}
	for next := ghUserSrc.NextUser(); next != first; next = ghUserSrc.NextUser() {
		got = append(got, next)
	}

	want := []string{
		"ajordens",
		"cfieber",
		"ethanfrogers",
		"ezimanyi",
		"jonsie",
		"mneterval@google.com",
		"plumpy@google.com",
		"robzienert",
		"ttomsu@google.com",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("did not get all expected users. want: %v, got %v", want, got)
	}
}
