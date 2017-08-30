package fb

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// TemplateContentType sets the content type of Flashback Template segments
	TemplateContentType = "text/html"
	// BundleContentType ses the content type of Flashback Bundles
	BundleContentType = "application/json"
)

// Bundle represents a Bundle database.
type Bundle struct {
	ID          string    `json:"_id"`
	Rev         string    `json:"_rev,omitempty"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	Imported    time.Time `json:"imported,omitempty"`
	Owner       string    `json:"owner"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Validate validates that all of the data in the bundle appears valid and self
// consistent. A nil return value means no errors were detected.
func (b *Bundle) Validate() error {
	if b.ID == "" {
		return errors.New("id required")
	}
	if err := validateDBID(b.ID); err != nil {
		return err
	}
	if !strings.HasPrefix(b.ID, "bundle-") {
		return errors.New("incorrect doc type")
	}
	if b.Created.IsZero() {
		return errors.New("created time required")
	}
	if b.Modified.IsZero() {
		return errors.New("modified time required")
	}
	if b.Owner == "" {
		return errors.New("owner required")
	}
	if _, err := B32dec(b.Owner); err != nil {
		return errors.Wrap(err, "invalid owner name")
	}
	return nil
}

// NewBundle creates a new Bundle with the provided id and owner.
func NewBundle(id, owner string) (*Bundle, error) {
	b := &Bundle{
		ID:       id,
		Owner:    owner,
		Created:  now().UTC(),
		Modified: now().UTC(),
	}
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return b, nil
}

type bundleAlias Bundle

// MarshalJSON implements the json.Marshaler interface for urnthe Bundle type.
func (b *Bundle) MarshalJSON() ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	doc := struct {
		bundleAlias
		Type     string     `json:"type"`
		Imported *time.Time `json:"imported,omitempty"`
	}{
		Type:        "bundle",
		bundleAlias: bundleAlias(*b),
	}
	if !b.Imported.IsZero() {
		doc.Imported = &b.Imported
	}
	return json.Marshal(doc)
}

// UnmarshalJSON fulfills the json.Unmarshaler interface for the Bundle type.
func (b *Bundle) UnmarshalJSON(data []byte) error {
	doc := &bundleAlias{}
	if err := json.Unmarshal(data, doc); err != nil {
		return errors.Wrap(err, "failed to unmarshal Bundle")
	}
	*b = Bundle(*doc)
	return b.Validate()
}

// SetRev sets the internal _rev attribute of the Bundle
func (b *Bundle) SetRev(rev string) { b.Rev = rev }

// DocID returns the document's ID as a string.
func (b *Bundle) DocID() string { return b.ID }

// ImportedTime returns the time the Bundle was imported, or nil
func (b *Bundle) ImportedTime() time.Time { return b.Imported }

// ModifiedTime returns the time the Bundle was last modified
func (b *Bundle) ModifiedTime() time.Time { return b.Modified }

// MergeImport attempts to merge i into b, returning true if a merge took place,
// or false if no merge was necessary.
func (b *Bundle) MergeImport(i interface{}) (bool, error) {
	existing := i.(*Bundle)
	if b.ID != existing.ID {
		return false, errors.New("IDs don't match")
	}
	if !b.Created.Equal(existing.Created) {
		return false, errors.New("Created timestamps don't match")
	}
	if b.Owner != existing.Owner {
		return false, errors.New("Cannot change bundle ownership")
	}
	if b.Imported.IsZero() || existing.Imported.IsZero() {
		return false, errors.New("not an import")
	}
	b.Rev = existing.Rev
	if b.Modified.After(existing.Modified) {
		// The new version is newer than the existing one, so update
		return true, nil
	}
	// The new version is older, so we need to use the version we just read
	b.Name = existing.Name
	b.Description = existing.Description
	b.Modified = existing.Modified
	b.Imported = existing.Imported
	return false, nil
}
