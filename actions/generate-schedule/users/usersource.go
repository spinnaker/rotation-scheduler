package users

import (
	"container/list"
	"fmt"
	"sort"
)

// Source represents a source of usernames for rotation
type Source interface {
	//GetUsers() ([]string, error)
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

//func (ss *StaticSource) GetUsers() ([]string, error) {
//	if ss.users.Front() == nil {
//		return nil, fmt.Errorf("empty user list")
//	}
//
//	users := make([]string, ss.users.Len())
//	for user := ss.users.Front(); user != nil; user.Next() {
//		users = append(users, user.Value.(string))
//	}
//	return users, nil
//}

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
