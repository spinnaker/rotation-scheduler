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
	ss := NewStaticSource("a", "b", "c")

	for _, tc := range []struct {
		desc       string
		startAfter string
		wantErr    bool
		next1      string
		next2      string
	}{
		{
			desc:       "start after a",
			startAfter: "a",
			wantErr:    false,
			next1:      "b",
			next2:      "c",
		},
		{
			desc:       "start after b",
			startAfter: "b",
			wantErr:    false,
			next1:      "c",
			next2:      "a",
		},
		{
			desc:       "start after c",
			startAfter: "c",
			wantErr:    false,
			next1:      "a",
			next2:      "b",
		},
		{
			desc:       "missing entry",
			startAfter: "missing",
			wantErr:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := ss.StartAfter(tc.startAfter)
			if tc.wantErr && err == nil {
				t.Errorf("err expected and not received.")
			} else if !tc.wantErr && err != nil {
				t.Errorf("got error from StartAfter: %v:", err)
			} else if tc.wantErr {
				// Successfully invoked error condition
				return
			}

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
