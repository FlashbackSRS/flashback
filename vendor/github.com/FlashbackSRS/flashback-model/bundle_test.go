package fb

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestNewBundle(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		owner    string
		expected *Bundle
		err      string
	}{
		{
			name:  "no id",
			owner: "user-mjxwe",
			err:   "id required",
		},
		{
			name: "no owner",
			id:   "bundle-mzxw6",
			err:  "owner required",
		},
		{
			name:  "invalid owner name",
			id:    "bundle-mzxw6",
			owner: "user-foo",
			err:   "invalid owner name: illegal base32 data at input byte 4",
		},
		{
			name:  "valid",
			id:    "bundle-mzxw6",
			owner: "mjxwe",
			expected: &Bundle{
				ID:       "bundle-mzxw6",
				Owner:    "mjxwe",
				Created:  now(),
				Modified: now(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewBundle(test.id, test.owner)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBundleMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bundle   *Bundle
		expected string
		err      string
	}{
		{
			name:   "fails validation",
			bundle: &Bundle{},
			err:    "id required",
		},
		{
			name: "null fields",
			bundle: &Bundle{
				ID:       "bundle-mzxw6",
				Owner:    "mjxwe",
				Created:  now(),
				Modified: now(),
			},
			expected: `{
				"_id":      "bundle-mzxw6",
				"type":     "bundle",
				"owner":    "mjxwe",
				"created":  "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z"
			}`,
		},
		{
			name: "all fields",
			bundle: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "mjxwe",
				Created:     now(),
				Modified:    now(),
				Imported:    now(),
				Name:        "foo name",
				Description: "foo description",
			},
			expected: `{
				"_id":         "bundle-mzxw6",
				"type":        "bundle",
				"owner":       "mjxwe",
				"name":        "foo name",
				"description": "foo description",
				"created":     "2017-01-01T00:00:00Z",
				"modified":    "2017-01-01T00:00:00Z",
				"imported":    "2017-01-01T00:00:00Z"
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.bundle.MarshalJSON()
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.JSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBundleUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Bundle
		err      string
	}{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   "failed to unmarshal Bundle: invalid character 'i' looking for beginning of value",
		},
		{
			name: "invalid user",
			input: `{
				"_id":      "bundle-mzxw6",
				"owner":    "unf",
				"created":  "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z"
			}`,
			err: "invalid owner name: illegal base32 data at input byte 3",
		},
		{
			name: "null fields",
			input: `{
				"_id":      "bundle-mzxw6",
				"owner":    "mjxwe",
				"created":  "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z"
			}`,
			expected: &Bundle{
				ID:       "bundle-mzxw6",
				Owner:    "mjxwe",
				Created:  now(),
				Modified: now(),
			},
		},
		{
			name: "all fields",
			input: `{
                "_id":         "bundle-mzxw6",
                "owner":       "mjxwe",
                "name":        "foo name",
                "description": "foo description",
                "created":     "2017-01-01T00:00:00Z",
                "modified":    "2017-01-01T00:00:00Z",
                "imported":    "2017-01-01T00:00:00Z"
            }`,
			expected: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "mjxwe",
				Created:     now(),
				Modified:    now(),
				Imported:    now(),
				Name:        "foo name",
				Description: "foo description",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Bundle{}
			err := result.UnmarshalJSON([]byte(test.input))
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBundleSetRev(t *testing.T) {
	bundle := &Bundle{}
	rev := "1-xxx"
	bundle.SetRev(rev)
	if bundle.Rev != rev {
		t.Errorf("failed to set rev")
	}
}

func TestBundleID(t *testing.T) {
	expected := "bundle-mzxw6"
	bundle := &Bundle{ID: "bundle-mzxw6"}
	if id := bundle.DocID(); id != expected {
		t.Errorf("unexpected id: %s", id)
	}
}

func TestBundleImportedTime(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		bundle := &Bundle{}
		ts := now()
		bundle.Imported = ts
		if it := bundle.ImportedTime(); it != ts {
			t.Errorf("Unexpected result: %s", it)
		}
	})
	t.Run("Unset", func(t *testing.T) {
		bundle := &Bundle{}
		if it := bundle.ImportedTime(); !it.IsZero() {
			t.Errorf("unexpected result: %v", it)
		}
	})
}

func TestBundleModifiedTime(t *testing.T) {
	bundle := &Bundle{}
	ts := now()
	bundle.Modified = ts
	if mt := bundle.ModifiedTime(); mt != ts {
		t.Errorf("Unexpected result")
	}
}

func TestBundleMergeImport(t *testing.T) {
	type Test struct {
		name           string
		new            *Bundle
		existing       *Bundle
		expected       bool
		expectedBundle *Bundle
		err            string
	}
	tests := []Test{
		{
			name:     "different ids",
			new:      &Bundle{ID: "bundle-mzxw6"},
			existing: &Bundle{ID: "bundle-mjqxecq"},
			err:      "IDs don't match",
		},
		{
			name:     "created timestamps don't match",
			new:      &Bundle{ID: "bundle-mzxw6", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Bundle{ID: "bundle-mzxw6", Created: parseTime("2017-02-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			err:      "Created timestamps don't match",
		},
		{
			name:     "owners don't match",
			new:      &Bundle{ID: "bundle-mzxw6", Owner: "user-mjxwe", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Bundle{ID: "bundle-mzxw6", Owner: "mfwgsy3fbi", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			err:      "Cannot change bundle ownership",
		},
		{
			name:     "new not an import",
			new:      &Bundle{ID: "bundle-mzxw6", Owner: "user-mjxwe", Created: parseTime("2017-01-01T01:01:01Z")},
			existing: &Bundle{ID: "bundle-mzxw6", Owner: "user-mjxwe", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			err:      "not an import",
		},
		{
			name:     "existing not an import",
			new:      &Bundle{ID: "bundle-mzxw6", Owner: "user-mjxwe", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Bundle{ID: "bundle-mzxw6", Owner: "user-mjxwe", Created: parseTime("2017-01-01T01:01:01Z")},
			err:      "not an import",
		},
		{
			name: "new is newer",
			new: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "user-mjxwe",
				Name:        "foo",
				Description: "FOO",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-15T00:00:00Z"),
			},
			existing: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "user-mjxwe",
				Name:        "bar",
				Description: "BAR",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-01-01T01:01:01Z"),
				Imported:    parseTime("2017-01-20T00:00:00Z"),
			},
			expected: true,
			expectedBundle: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "user-mjxwe",
				Name:        "foo",
				Description: "FOO",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-15T00:00:00Z"),
			},
		},
		{
			name: "existing is newer",
			new: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "user-mjxwe",
				Name:        "foo",
				Description: "FOO",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-01-01T01:01:01Z"),
				Imported:    parseTime("2017-01-15T00:00:00Z"),
			},
			existing: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "user-mjxwe",
				Name:        "bar",
				Description: "BAR",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-20T00:00:00Z"),
			},
			expected: false,
			expectedBundle: &Bundle{
				ID:          "bundle-mzxw6",
				Owner:       "user-mjxwe",
				Name:        "bar",
				Description: "BAR",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-20T00:00:00Z"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.new.MergeImport(test.existing)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if test.expected != result {
				t.Errorf("Unexpected result: %t", result)
			}
			if d := diff.Interface(test.expectedBundle, test.new); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBundleValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "no ID",
			v:    &Bundle{},
			err:  "id required",
		},
		{
			name: "invalid doctype",
			v:    &Bundle{ID: "chicken-mzxw6"},
			err:  "unsupported DBID type 'chicken'",
		},
		{
			name: "wrong doctype",
			v:    &Bundle{ID: "user-mzxw6"},
			err:  "incorrect doc type",
		},
		{
			name: "no created time",
			v:    &Bundle{ID: "bundle-mzxw6"},
			err:  "created time required",
		},
		{
			name: "no modified time",
			v:    &Bundle{ID: "bundle-mzxw6", Created: now()},
			err:  "modified time required",
		},
		{
			name: "no owner",
			v:    &Bundle{ID: "bundle-mzxw6", Created: now(), Modified: now()},
			err:  "owner required",
		},
		{
			name: "invalid user",
			v:    &Bundle{ID: "bundle-mzxw6", Owner: "foo-bar", Created: now(), Modified: now()},
			err:  "invalid owner name: illegal base32 data at input byte 3",
		},
		{
			name: "valid",
			v:    &Bundle{ID: "bundle-mzxw6", Owner: "mjxwe", Created: now(), Modified: now()},
		},
	}
	testValidation(t, tests)
}
