package fb

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// Attachment represents a Couch/PouchDB attachment.
type Attachment struct {
	ContentType string `json:"content_type"`
	Content     []byte `json:"data"`
}

// FileCollection represents a collection of Attachments which may be used by
// multiple related sub-document elements.
type FileCollection struct {
	files map[string]*Attachment
	views []*FileCollectionView
}

// FileList returns a list of filenames contained within the collection.
func (fc *FileCollection) FileList() []string {
	files := make([]string, 0, len(fc.files))
	for name := range fc.files {
		files = append(files, name)
	}
	return files
}

// GetFile returns an Attachment based on the file name. If the file does not
// the second return value will be false.
func (fc *FileCollection) GetFile(name string) (*Attachment, bool) {
	att, ok := fc.files[name]
	return att, ok
}

// FileCollectionView represents a view into a larger FileCollection, which can
// be used by sub-elements.
type FileCollectionView struct {
	col     *FileCollection
	members map[string]*Attachment
}

// NewFileCollection returns a new, empty FileCollection.
func NewFileCollection() *FileCollection {
	return &FileCollection{
		files: make(map[string]*Attachment),
		views: make([]*FileCollectionView, 0, 1),
	}
}

// AddView creates a new View on a FileCollection, which can be used by sub-elements.
func (fc *FileCollection) AddView(v *FileCollectionView) error {
	for filename := range v.members {
		att, ok := fc.files[filename]
		if !ok {
			return errors.New(filename + " not found in collection")
		}
		v.members[filename] = att
	}
	v.col = fc
	fc.views = append(fc.views, v)
	return nil
}

// RemoveView removes a FileCollectionView from a FileCollection
func (fc *FileCollection) RemoveView(v *FileCollectionView) error {
	for filename := range v.members {
		delete(fc.files, filename)
	}
	for i, view := range fc.views {
		if view == v {
			fc.views = append(fc.views[:i], fc.views[i+1:]...)
			return nil
		}
	}
	return errors.New("view not found")
}

// NewView returns a new FileCollectionView from the existing FileCollection.
func (fc *FileCollection) NewView() *FileCollectionView {
	v := &FileCollectionView{
		col:     fc,
		members: make(map[string]*Attachment),
	}
	fc.views = append(fc.views, v)
	return v
}

// RemoveFile removes all references to the named Attachment.
func (fc *FileCollection) RemoveFile(name string) {
	delete(fc.files, name)
	for _, view := range fc.views {
		delete(view.members, name)
	}
}

// filenameEscapeChar is used to escape the first character of attachments
// that begin with '_' or the escape char itself. '^' was chosen, as it does
// not require special JSON escaping (as '\' would, for example), and should
// appear rarely in filenames.
const filenameEscapeChar = '^'

// EscapeFilename and UnescapeFilename convert filenames to legal PouchDB
// representations. In particular, this means non-ASCII and special characters
// are URL-encoded, and leading '_' characters as well, as these upset PouchDB.
// Any '_' characters found elsewhere in the filename are left alone, to
// preserve a few bytes of space (woot!).
func EscapeFilename(filename string) string {
	if len(filename) == 0 {
		return filename
	}
	switch filename[0] {
	case '_', filenameEscapeChar:
		return string(filenameEscapeChar) + filename
	}
	return filename
}

// UnescapeFilename converts a filename from its PouchDB reprsentation to its
// original form.
func UnescapeFilename(escaped string) string {
	return strings.TrimPrefix(escaped, string(filenameEscapeChar))
}

// MarshalJSON implements the json.Marshaler interface for the FileCollection type.
func (fc *FileCollection) MarshalJSON() ([]byte, error) {
	escaped := make(map[string]*Attachment)
	for filename, attachment := range fc.files {
		escaped[EscapeFilename(filename)] = attachment
	}
	return json.Marshal(escaped)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the FileCollection type.
func (fc *FileCollection) UnmarshalJSON(data []byte) error {
	escaped := make(map[string]*Attachment)
	if err := json.Unmarshal(data, &escaped); err != nil {
		return err
	}
	fc.files = make(map[string]*Attachment)
	fc.views = make([]*FileCollectionView, 0)
	for escapedName, attachment := range escaped {
		filename := UnescapeFilename(escapedName)
		fc.files[filename] = attachment
	}
	return nil
}

// hasMemberView returns true if view is a member of fc.
func (fc *FileCollection) hasMemberView(view *FileCollectionView) bool {
	for _, v := range fc.views {
		if view == v {
			return true
		}
	}
	return false
}

// SetFile sets the requested attachment, replacing it if it already exists.
func (v *FileCollectionView) SetFile(name, ctype string, content []byte) {
	att := &Attachment{
		ContentType: ctype,
		Content:     content,
	}
	v.col.files[name] = att
	v.members[name] = att
}

// AddFile adds the requested attachment. Returns an error if it already exists.
func (v *FileCollectionView) AddFile(name, ctype string, content []byte) error {
	if _, ok := v.col.files[name]; ok {
		return errors.Errorf("'%s' already exists in the collection", name)
	}
	v.SetFile(name, ctype, content)
	return nil
}

// RemoveFile removes the named attachment from the collection.
func (v *FileCollectionView) RemoveFile(name string) error {
	if _, ok := v.members[name]; !ok {
		return errors.New("file not found in view")
	}
	delete(v.members, name)
	v.col.RemoveFile(name)
	return nil
}

// FileList returns a list of filenames contained within the view.
func (v *FileCollectionView) FileList() []string {
	files := make([]string, 0, len(v.members))
	for name := range v.members {
		files = append(files, name)
	}
	return files
}

// GetFile returns an Attachment based on the file name. If the file does not
// the second return value will be false.
func (v *FileCollectionView) GetFile(name string) (*Attachment, bool) {
	att, ok := v.members[name]
	return att, ok
}

// MarshalJSON implements the json.Marshaler interface for the FileCollectionView type.
func (v *FileCollectionView) MarshalJSON() ([]byte, error) {
	names := make([]string, 0, len(v.members))
	for name := range v.members {
		names = append(names, name)
	}
	sort.Strings(names) // For consistent output
	return json.Marshal(names)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the FileCollectionView type.
func (v *FileCollectionView) UnmarshalJSON(data []byte) error {
	v.members = make(map[string]*Attachment)
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return errors.Wrap(err, "failed to unmarshal file collection view")
	}
	for _, filename := range names {
		v.members[filename] = nil
	}
	return nil
}
