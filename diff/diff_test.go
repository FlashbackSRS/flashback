package diff

import "testing"

func TestDiff(t *testing.T) {
	tests := []struct {
		name         string
		text1, text2 string
		equal        bool
		diff         string
	}{
		{
			name:  "equal",
			text1: "foo",
			text2: "foo",
			equal: true,
			diff:  `<span class="good">foo</span>`,
		},
		{
			name:  "not equal",
			text1: "enseigner",
			text2: "ensignier",
			equal: false,
			diff:  `<span class="good">ens</span><span class="del">e</span><span class="good">ign</span><span class="ins">i</span><span class="good">er</span>`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			equal, diff := Diff(test.text1, test.text2)
			if test.equal != equal {
				t.Errorf("Unexpected equality: %t\n", equal)
			}
			if test.diff != diff {
				t.Errorf("Unexpected diff: %s\n", diff)
			}
		})
	}
}
