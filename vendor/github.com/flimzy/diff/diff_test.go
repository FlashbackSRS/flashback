package diff

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestResultString(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var r *Result
		expected := ""
		if result := r.String(); result != expected {
			t.Errorf("Unexpected result: %s", result)
		}
	})
	t.Run("diff", func(t *testing.T) {
		expected := "foo"
		r := &Result{diff: expected}
		if result := r.String(); result != expected {
			t.Errorf("Unexpected result: %s", result)
		}
	})
}

func TestSliceDiff(t *testing.T) {
	tests := []struct {
		name             string
		expected, actual []string
		result           string
	}{
		{
			name:     "equal",
			expected: []string{"foo"},
			actual:   []string{"foo"},
		},
		{
			name:     "different",
			expected: []string{"foo"},
			actual:   []string{"bar"},
			result:   "--- expected\n+++ actual\n@@ -1 +1 @@\n-foo+bar",
		},
	}
	for _, test := range tests {
		result := sliceDiff(test.expected, test.actual)
		var resultText string
		if result != nil {
			resultText = result.String()
		}
		if resultText != test.result {
			t.Errorf("Unexpected result:\n%s\n", resultText)
		}
	}
}

func TestTextSlices(t *testing.T) {
	tests := []struct {
		name             string
		expected, actual []string
		result           string
	}{
		{
			name:     "equal",
			expected: []string{"foo", "bar"},
			actual:   []string{"foo", "bar"},
		},
		{
			name:     "different",
			expected: []string{"foo", "bar"},
			actual:   []string{"bar", "bar"},
			result:   "--- expected\n+++ actual\n@@ -1,2 +1,2 @@\n-foo\n bar\n+bar\n",
		},
	}
	for _, test := range tests {
		result := TextSlices(test.expected, test.actual)
		var resultText string
		if result != nil {
			resultText = result.String()
		}
		if resultText != test.result {
			t.Errorf("Unexpected result:\n%s\n", resultText)
		}
	}
}

func TestText(t *testing.T) {
	tests := []struct {
		name             string
		expected, actual string
		result           string
	}{
		{
			name:     "equal",
			expected: "foo\nbar\n",
			actual:   "foo\nbar\n",
		},
		{
			name:     "different",
			expected: "foo\nbar",
			actual:   "bar\nbar",
			result:   "--- expected\n+++ actual\n@@ -1,2 +1,2 @@\n-foo\n bar\n+bar\n",
		},
	}
	for _, test := range tests {
		result := Text(test.expected, test.actual)
		var resultText string
		if result != nil {
			resultText = result.String()
		}
		if resultText != test.result {
			t.Errorf("Unexpected result:\n%s\n", resultText)
		}
	}
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		isJSON bool
		result string
	}{
		{
			name:   "io.Reader",
			input:  strings.NewReader("foo"),
			isJSON: true,
			result: "foo",
		},
		{
			name:   "byte slice",
			input:  []byte("foo"),
			isJSON: true,
			result: "foo",
		},
		{
			name:   "json.RawMessage",
			input:  json.RawMessage("foo"),
			isJSON: true,
			result: "foo",
		},
		{
			name:   "string",
			input:  "foo",
			isJSON: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isJSON, result, _ := isJSON(test.input)
			if isJSON != test.isJSON {
				t.Errorf("Unexpected result: %t", isJSON)
			}
			if string(result) != test.result {
				t.Errorf("Unexpected result: %s", string(result))
			}
		})
	}
}

type errorReader struct{}

var _ io.Reader = &errorReader{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		err      string
	}{
		{
			name:     "byte slice",
			input:    []byte(`"foo"`),
			expected: `"foo"`,
		},
		{
			name:     "string",
			input:    "foo",
			expected: `"foo"`,
		},
		{
			name:  "invalid json",
			input: []byte("invalid json"),
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "error reader",
			input: &errorReader{},
			err:   "read error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := marshal(test.input)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.err != errMsg {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if string(result) != test.expected {
				t.Errorf("Unexpected result: %s", string(result))
			}
		})
	}
}

func TestAsJSON(t *testing.T) {
	tests := []struct {
		name             string
		expected, actual interface{}
		result           string
	}{
		{
			name:     "equal",
			expected: []string{"foo", "bar"},
			actual:   []string{"foo", "bar"},
		},
		{
			name:     "different",
			expected: []string{"foo", "bar"},
			actual:   []string{"bar", "bar"},
			result: `--- expected
+++ actual
@@ -1,4 +1,4 @@
 [
-    "foo",
+    "bar",
     "bar"
 ]
`,
		},
		{
			name:     "Unmarshalable expected",
			expected: make(chan int),
			actual:   "foo",
			result:   "failed to marshal expected value: json: unsupported type: chan int",
		},
		{
			name:     "Unmarshalable actual",
			expected: "foo",
			actual:   make(chan int),
			result:   "failed to marshal actual value: json: unsupported type: chan int",
		},
		{
			name:     "empty reader",
			expected: strings.NewReader(""),
			actual:   nil,
			result:   "",
		},
	}
	for _, test := range tests {
		result := AsJSON(test.expected, test.actual)
		var resultText string
		if result != nil {
			resultText = result.String()
		}
		if resultText != test.result {
			t.Errorf("Unexpected result:\n%s\n", resultText)
		}
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name             string
		expected, actual string
		result           string
	}{
		{
			name:     "equal",
			expected: `["foo","bar"]`,
			actual:   `["foo","bar"]`,
		},
		{
			name:     "different",
			expected: `["foo","bar"]`,
			actual:   `["bar","bar"]`,
			result: `--- expected
+++ actual
@@ -1,4 +1,4 @@
 [
-    "foo",
+    "bar",
     "bar"
 ]
`,
		},
		{
			name:     "invalid expected",
			expected: "invalid json",
			actual:   `"foo"`,
			result:   "failed to unmarshal expected value: invalid character 'i' looking for beginning of value",
		},
		{
			name:     "invalid actual",
			expected: `"foo"`,
			actual:   "invalid json",
			result:   "failed to unmarshal actual value: invalid character 'i' looking for beginning of value",
		},
		{
			name:     "empty",
			expected: "",
			actual:   "",
		},
	}
	for _, test := range tests {
		result := JSON([]byte(test.expected), []byte(test.actual))
		var resultText string
		if result != nil {
			resultText = result.String()
		}
		if resultText != test.result {
			t.Errorf("Unexpected result:\n%s\n", resultText)
		}
	}
}

func TestInterface(t *testing.T) {
	tests := []struct {
		name             string
		expected, actual interface{}
		result           string
	}{
		{
			name:     "equal",
			expected: []string{"foo", "bar"},
			actual:   []string{"foo", "bar"},
		},
		{
			name:     "different",
			expected: []string{"foo", "bar"},
			actual:   []string{"bar", "bar"},
			result: `--- expected
+++ actual
@@ -1,4 +1,4 @@
 ([]string) (len=2) {
-  (string) (len=3) "foo",
+  (string) (len=3) "bar",
   (string) (len=3) "bar"
 }
`,
		},
	}
	for _, test := range tests {
		result := Interface(test.expected, test.actual)
		var resultText string
		if result != nil {
			resultText = result.String()
		}
		if resultText != test.result {
			t.Errorf("Unexpected result:\n%s\n", resultText)
		}
	}
}
