package anki

import (
	"html/template"
	"testing"
)

func TestImage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected template.HTML
	}{
		{
			name:     "normal",
			filename: "foo.jpg",
			expected: `<img src="foo.jpg">`,
		},
		{
			name:     "escaped chars",
			filename: `foo"bar.jpg`,
			expected: `<img src="foo%22bar.jpg">`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := image(test.filename)
			if test.expected != result {
				t.Errorf("Expected: %s\n  Actual: %s\n", test.expected, result)
			}
		})
	}
}

func TestAudio(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		ctype    string
		expected template.HTML
	}{
		{
			name:     "normal",
			filename: "foo.mp3",
			ctype:    "audo/mpeg",
			expected: `<audio src="foo.mp3" type="audo/mpeg"></audio>`,
		},
		{
			name:     "escaped chars",
			filename: `foo"bar.mp3`,
			ctype:    "audo/mpeg",
			expected: `<audio src="foo%22bar.mp3" type="audo/mpeg"></audio>`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := audio(test.filename, test.ctype)
			if test.expected != result {
				t.Errorf("Expected: %s\n  Actual: %s\n", test.expected, result)
			}
		})
	}
}
