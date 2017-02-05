package diff

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Diff finds the differences between two strings, returned as HTML, and a
// boolean indicating if the strings are the same.
func Diff(text1, text2 string) (equal bool, diff string) {
	text1 = strings.TrimSpace(text1)
	text2 = strings.TrimSpace(text2)
	if text1 == text2 {
		return true, fmt.Sprintf("<spam class=\"good\">%s</span>", text1)
	}
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)
	return false, PrettyHTML(dmp, diffs)
}

// PrettyHTML produces pretty HTML for display.
//
// This function is copied and modified from the function `DiffPrettyHtml()` in
// the github.com/sergi/go-diff/diffmatchpatch package.
func PrettyHTML(dmp *diffmatchpatch.DiffMatchPatch, diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := strings.Replace(html.EscapeString(diff.Text), "\n", "&para;<br>", -1)
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("<span class=\"ins\">")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</span>")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("<span class=\"del\">")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</span>")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("<span class=\"good\">")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</span>")
		}
	}
	return buff.String()
}
