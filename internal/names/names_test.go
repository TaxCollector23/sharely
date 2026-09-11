package names

import "testing"

func TestGenerateAvoidsTaken(t *testing.T) {
	taken := map[string]bool{}
	for i := 0; i < len(words)+5; i++ {
		name, err := Generate(taken)
		if err != nil {
			t.Fatal(err)
		}
		if taken[name] {
			t.Fatalf("Generate returned an already-taken name: %s", name)
		}
		taken[name] = true
	}
}

func TestGenerateIsReadable(t *testing.T) {
	name, err := Generate(map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if len(name) == 0 {
		t.Fatal("expected a non-empty name")
	}
}
