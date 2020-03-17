package users

import (
	"container/list"
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
	users    *list.List
	nextUser *list.Element
}

func NewStaticSource(users ...string) *StaticSource {
	sort.Strings(users)
	ss := &StaticSource{users: list.New()}
	for _, u := range users {
		ss.users.PushBack(u)
	}
	ss.nextUser = ss.users.Front()
	return ss
}

func (ss *StaticSource) StartAfter(user string) error {
	for c := ss.users.Front(); c != nil; c = c.Next() {
		if user == c.Value.(string) {
			ss.nextUser = c
			ss.NextUser()
			return nil
		}
	}

	return fmt.Errorf("cannot start after user %v, user not found", user)
}

func (ss *StaticSource) NextUser() string {
	returnVal := ss.nextUser.Value.(string)

	ss.nextUser = ss.nextUser.Next()
	if ss.nextUser == nil {
		ss.nextUser = ss.users.Front()
	}

	return returnVal
}
