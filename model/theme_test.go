package model

import (
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
)

func TestExtractTemplateFiles(t *testing.T) {
	type etfTest struct {
		name     string
		view     *fb.FileCollectionView
		expected map[string]string
		err      string
	}
	tests := []etfTest{
		{
			name:     "no files",
			view:     &fb.FileCollectionView{},
			expected: map[string]string{},
		},
		{
			name: "template files",
			view: func() *fb.FileCollectionView {
				c := fb.NewFileCollection()
				v := c.NewView()
				_ = v.AddFile("foo.txt", "text/plain", []byte("text"))
				_ = v.AddFile("template", fb.TemplateContentType, []byte("template content"))
				_ = v.AddFile("javascript", "script/javascript", []byte("js content"))
				_ = v.AddFile("css", "text/css", []byte("css content"))
				return v
			}(),
			expected: map[string]string{
				"template":   "template content",
				"css":        "css content",
				"javascript": "js content",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractTemplateFiles(test.view)
			if d := diff.Interface(test.expected, result); d != "" {
				t.Error(d)
			}
		})
	}
}
