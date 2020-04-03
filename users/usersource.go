package users

import (
	"container/ring"
	"fmt"
	"sort"
)

// Source represents a source of usernames for rotation
type Source interface {
	StartAfter(user string) error
	NextUser() string
}

// StaticSource is a base implementation of static list of usernames.
type StaticSource struct {
	nextUser *ring.Ring
}

func NewStaticSource(users ...string) *StaticSource {
	sort.Strings(users)
	ss := &StaticSource{nextUser: ring.New(len(users))}
	for _, u := range users {
		ss.nextUser.Value = u
		ss.nextUser = ss.nextUser.Next()
	}
	return ss
}

func (ss *StaticSource) StartAfter(user string) error {
	for linksChecked := 0; linksChecked <= ss.nextUser.Len(); linksChecked++ {
		if user == ss.nextUser.Prev().Value.(string) {
			// User found, and ring is currently in correct state.
			return nil
		}
		ss.nextUser = ss.nextUser.Next()
	}

	return fmt.Errorf("cannot start after user %v, user not found", user)
}

func (ss *StaticSource) NextUser() string {
	u := ss.nextUser.Value.(string)
	ss.nextUser = ss.nextUser.Next()
	return u
}
