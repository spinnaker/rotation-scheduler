package users

import (
	"container/ring"
	"sort"
	"strings"
)

// Source represents a source of usernames for rotation.
type Source interface {
	// StartAfter positions this Source to begin iterating alphabetically after this user. If this user sorts (case
	// insensitively) outside of the first or last user, the first user is returned.
	StartAfter(user string)
	NextUser() string
}

// StaticSource is a base implementation of static list of usernames.
type StaticSource struct {
	nextUser *ring.Ring
}

// NewStaticSource creates a StaticSource of all lower-cased and sorted from the specified users.
func NewStaticSource(users ...string) *StaticSource {
	for i, u := range users {
		users[i] = strings.ToLower(u)
	}
	sort.Strings(users)

	ss := &StaticSource{nextUser: ring.New(len(users))}
	for _, u := range users {
		ss.nextUser.Value = u
		ss.nextUser = ss.nextUser.Next()
	}
	return ss
}

func (ss *StaticSource) StartAfter(user string) {
	user = strings.ToLower(user)

	var beginning *ring.Ring
	for linksChecked := 0; linksChecked <= ss.nextUser.Len(); linksChecked++ {
		prevUser := ss.nextUser.Prev().Value.(string)
		nextUser := ss.nextUser.Value.(string)
		if nextUser < prevUser {
			// We need the beginning of the cycle in case the user doesn't fall in between any of the existing users.
			beginning = ss.nextUser
		}
		if prevUser <= user && user < nextUser {
			// User found, and ring is currently in correct state.
			return
		}
		ss.nextUser = ss.nextUser.Next()
	}
	ss.nextUser = beginning
}

func (ss *StaticSource) NextUser() string {
	u := ss.nextUser.Value.(string)
	ss.nextUser = ss.nextUser.Next()
	return u
}
