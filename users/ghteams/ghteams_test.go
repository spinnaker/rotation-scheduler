package ghteams

import (
	"context"
	"reflect"
	"testing"

	"cloud.google.com/go/httpreplay"
)

func TestNewGitHubTeamsUserSource(t *testing.T) {
	r, err := httpreplay.NewReplayer("testing/teams.replay")
	if err != nil {
		t.Fatalf("error creating replayer: %v", err)
	}

	client, err := r.Client(context.Background())
	if err != nil {
		t.Fatalf("error creating replayer client: %v", err)
	}

	ghUserSrc, err := NewGitHubTeamsUserSource(client, "spinnaker", "build-cops")
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
		"duftler",
		"ethanfrogers",
		"ezimanyi",
		"jonsie",
		"maggieneterval",
		"plumpy",
		"robzienert",
		"ttomsu",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("did not get all expected users. want: %v, got %v", want, got)
	}
}
