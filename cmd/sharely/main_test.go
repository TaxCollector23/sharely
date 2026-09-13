package main

import "testing"

func TestParseShareFlags(t *testing.T) {
	f, err := parseShareFlags([]string{"./photos", "--expires", "15minutes", "--password", "--port", "9000"})
	if err != nil {
		t.Fatal(err)
	}
	if f.target != "./photos" || f.expires != "15m" || !f.password || f.port != 9000 {
		t.Fatalf("unexpected flags: %#v", f)
	}
}

func TestParseShareFlagsRejectsBadInput(t *testing.T) {
	tests := [][]string{
		{"--expires", "tomorrow"},
		{"--port", "nope"},
		{"--host", "example.com"},
		{"--unknown"},
		{"one", "two"},
	}
	for _, args := range tests {
		if _, err := parseShareFlags(args); err == nil {
			t.Errorf("parseShareFlags(%q) should fail", args)
		}
	}
}

func TestNormalizeExpires(t *testing.T) {
	cases := map[string]string{"15min": "15m", "hour": "1h", "4hours": "4h", "until-stopped": "forever"}
	for input, want := range cases {
		if got := normalizeExpires(input); got != want {
			t.Errorf("normalizeExpires(%q) = %q, want %q", input, got, want)
		}
	}
}
