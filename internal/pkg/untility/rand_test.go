package untility

import "testing"

func TestRandStr(t *testing.T) {
	got := len(RandStr(4))
	want := 4

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

}
