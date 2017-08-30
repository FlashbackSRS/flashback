package fb

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestNewModel(t *testing.T) {
	type Test struct {
		name      string
		theme     *Theme
		modelType string
		expected  *Model
		err       string
	}
	tests := []Test{
		{
			name: "nil theme",
			err:  "theme is required",
		},
		{
			name:  "no type",
			theme: &Theme{},
			err:   "model type is required",
		},
		{
			name: "valid",
			theme: func() *Theme {
				theme, _ := NewTheme("theme-foo")
				return theme
			}(),
			modelType: "foo",
			expected: func() *Model {
				theme, _ := NewTheme("theme-foo")
				// att := NewFileCollection()
				// theme.Files = att.NewView()
				theme.ModelSequence = 1
				model := &Model{
					Type:      "foo",
					Templates: []string{},
					Fields:    []*Field{},
					Files:     theme.Attachments.NewView(),
					Theme:     theme,
				}
				return model
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewModel(test.theme, test.modelType)
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

func TestModelValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "nil theme",
			v:    &Model{},
			err:  "theme is required",
		},
		{
			name: "empty type",
			v:    &Model{Theme: &Theme{}},
			err:  "type is required",
		},
		{
			name: "nil files",
			v:    &Model{Theme: &Theme{}, Type: "foo"},
			err:  "file list must not be nil",
		},
		{
			name: "incomplete theme",
			v:    &Model{Theme: &Theme{}, Type: "foo", Files: NewFileCollection().NewView()},
			err:  "invalid theme",
		},
		{
			name: "valid",
			v: func() *Model {
				att := NewFileCollection()
				return &Model{Theme: &Theme{Attachments: att}, Type: "foo", Files: att.NewView()}
			}(),
		},
	}
	testValidation(t, tests)
}

func TestModelAddfile(t *testing.T) {
	type Test struct {
		name     string
		model    *Model
		filename string
		expected interface{}
		err      string
	}
	tests := []Test{
		{
			name: "duplicate",
			model: func() *Model {
				theme, _ := NewTheme("theme-foo")
				model := &Model{
					Theme: theme,
					Files: theme.Attachments.NewView(),
				}
				_ = model.AddFile("foo.txt", "text/plain", []byte("foo"))
				theme.Models = []*Model{model}
				return model
			}(),
			filename: "foo.txt",
			err:      "'foo.txt' already exists in the collection",
		},
		{
			name: "success",
			model: func() *Model {
				theme, _ := NewTheme("theme-Zm9v")
				model := &Model{
					Type:  "test",
					Theme: theme,
					Files: theme.Attachments.NewView(),
				}
				theme.Models = []*Model{model}
				theme.ModelSequence = 1
				return model
			}(),
			filename: "foo.txt",
			expected: map[string]interface{}{
				"_id":      "theme-Zm9v",
				"type":     "theme",
				"created":  now(),
				"modified": now(),
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "Zm9v",
					},
				},
				"files":         []string{},
				"modelSequence": 1,
				"models": []map[string]interface{}{
					{
						"id":        0,
						"modelType": "test",
						"templates": nil,
						"fields":    nil,
						"files":     []string{"foo.txt"},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.model.AddFile(test.filename, "text/plain", []byte("foo"))
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, test.model.Theme); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestModelIdentity(t *testing.T) {
	t.Run("Null theme", func(t *testing.T) {
		model := &Model{}
		expected := ""
		if id := model.Identity(); id != expected {
			t.Errorf("Unexpected result: %s", id)
		}
	})
	t.Run("full id", func(t *testing.T) {
		model := &Model{
			Theme: &Theme{ID: "theme-abcd"},
			ID:    1,
		}
		expected := "abcd.1"
		if id := model.Identity(); id != expected {
			t.Errorf("Unexpected result: %s", id)
		}
	})
}

func TestModelAddField(t *testing.T) {
	type Test struct {
		name     string
		model    *Model
		fType    FieldType
		fName    string
		expected interface{}
		err      string
	}
	tests := []Test{
		{
			name:  "invalid type",
			fType: 9999,
			err:   "invalid field type",
		},
		{
			name:  "missing name",
			fType: 1,
			err:   "field name is required",
		},
		{
			name:  "valid",
			fType: 1,
			fName: "Foo",
			model: &Model{Fields: []*Field{}},
			expected: map[string]interface{}{
				"id":        0,
				"modelType": "",
				"templates": nil,
				"fields": []map[string]interface{}{
					{
						"fieldType": 1,
						"name":      "Foo",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.model.AddField(test.fType, test.fName)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, test.model); d != nil {
				t.Error(d)
			}
		})
	}
}
