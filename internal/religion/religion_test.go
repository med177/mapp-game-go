package religion

import "testing"

func TestDisplayNameTR(t *testing.T) {
	if got := DisplayNameTR(Sunni); got != "Sünni İslam" {
		t.Fatalf("display name mismatch: got=%q", got)
	}
}

func TestNext(t *testing.T) {
	if got := Next(Orthodox); got != Sunni {
		t.Fatalf("next religion mismatch: got=%q want=%q", got, Sunni)
	}
}
