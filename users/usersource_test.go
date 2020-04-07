package users

import "testing"

func TestNextUser(t *testing.T) {
	ss := NewStaticSource("a", "b")
	got := ss.NextUser()
	if "a" != got {
		t.Errorf("(first pass) want a, got %v", got)
	}

	got = ss.NextUser()
	if "b" != got {
		t.Errorf("want b, got %v", got)
	}

	got = ss.NextUser()
	if "a" != got {
		t.Errorf("(second pass) want a, got %v", got)
	}
}

func TestStartAfter(t *testing.T) {
	for _, tc := range []struct {
		desc       string
		users      []string
		startAfter string
		next1      string
		next2      string
	}{
		{
			desc:       "start after missing, single element",
			users:      []string{"a"},
			startAfter: "missing",
			next1:      "a",
			next2:      "a",
		},
		{
			desc:       "start after a",
			users:      []string{"a", "b", "c"},
			startAfter: "a",
			next1:      "b",
			next2:      "c",
		},
		{
			desc:       "start after b",
			users:      []string{"a", "b", "c"},
			startAfter: "b",
			next1:      "c",
			next2:      "a",
		},
		{
			desc:       "start after c",
			users:      []string{"a", "b", "c"},
			startAfter: "c",
			next1:      "a",
			next2:      "b",
		},
		{
			desc:       "unsorted mixed cases",
			users:      []string{"C", "b", "A"},
			startAfter: "a",
			next1:      "b",
			next2:      "c",
		},
		{
			desc:       "missing entry easy next",
			users:      []string{"a", "c", "d"},
			startAfter: "b",
			next1:      "c",
			next2:      "d",
		},
		{
			desc:       "missing entry at front",
			users:      []string{"b", "c", "d"},
			startAfter: "a",
			next1:      "b",
			next2:      "c",
		},
		{
			desc:       "missing entry at back",
			users:      []string{"b", "c", "d"},
			startAfter: "z",
			next1:      "b",
			next2:      "c",
		},
		{
			desc:       "non-letter startAfter",
			users:      []string{"b", "c", "d"},
			startAfter: "123",
			next1:      "b",
			next2:      "c",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			ss := NewStaticSource(tc.users...)

			ss.StartAfter(tc.startAfter)

			got := ss.NextUser()
			if tc.next1 != got {
				t.Errorf("next1: want %v, got %v", tc.next1, got)
			}
			got = ss.NextUser()
			if tc.next2 != got {
				t.Errorf("next2: want %v, got %v", tc.next2, got)
			}
		})
	}
}

func TestContains(t *testing.T) {
	ss := NewStaticSource("a")

	if !ss.Contains("a") {
		t.Errorf("should contain a")
	} else if ss.Contains("b") {
		t.Errorf("should not contain b")
	}
}
