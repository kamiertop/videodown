package settings

import "testing"

func TestReorderRepeatedFilenameField(t *testing.T) {
	s := NewState()
	s.Config.Filename = "{title} - {author} - {title} - {id}"
	s.moveFilenamePart(filenamePart{Token: "{title}", Occurrence: 1}, 4)
	if s.Config.Filename != "{title} - {author} - {id} - {title}" {
		t.Fatalf("moved the wrong occurrence: %s", s.Config.Filename)
	}
	s.moveFilenamePart(filenamePart{Token: "{missing}"}, 0)
	if s.Config.Filename != "{title} - {author} - {id} - {title}" {
		t.Fatal("unknown field changed the template")
	}
}
