package mimekind

import "testing"

func TestDetect(t *testing.T) {
	cases := []struct {
		name string
		cat  Category
	}{
		{"index.html", CategoryHTML},
		{"report.pdf", CategoryPDF},
		{"photo.png", CategoryImage},
		{"photo.jpg", CategoryImage},
		{"clip.mp4", CategoryVideo},
		{"song.mp3", CategoryAudio},
		{"notes.md", CategoryText},
		{"data.json", CategoryText},
		{"archive.zip", CategoryOther},
	}
	for _, c := range cases {
		_, got := Detect(c.name)
		if got != c.cat {
			t.Errorf("Detect(%q) category = %s, want %s", c.name, got, c.cat)
		}
	}
}
