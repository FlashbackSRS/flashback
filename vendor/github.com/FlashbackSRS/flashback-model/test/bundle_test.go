package test

import (
	"encoding/json"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback-model"
)

var frozenBundle = []byte(`
{
	"_id": "bundle-krsxg5baij2w4zdmmu",
	"type": "bundle",
	"created": "2016-07-31T15:08:24.730156517Z",
	"modified": "2016-07-31T15:08:24.730156517Z",
	"imported": "2016-08-02T15:08:24.730156517Z",
	"owner": "tui5ajfbabaeljnxt4om7fwmt4",
	"name": "Test Bundle",
	"description": "A bundle for testing"
}
`)

func TestNewBundle(t *testing.T) {
	require := require.New(t)
	b, err := fb.NewBundle("bundle-krsxg5baij2w4zdmmu", "tui5ajfbabaeljnxt4om7fwmt4")
	require.Nil(err, "Error creating new bundle: %s", err)

	b.Name = "Test Bundle"
	b.Created = now
	b.Modified = now
	b.Imported = now.AddDate(0, 0, 2)
	b.Description = "A bundle for testing"
	require.Equal("bundle-krsxg5baij2w4zdmmu", b.ID, "Bundle ID")
	require.MarshalsToJSON(frozenBundle, b, "New Bundle")

	b2 := &fb.Bundle{}
	err = json.Unmarshal(frozenBundle, b2)
	require.Nil(err, "Error thawing bundle: %s", err)
	require.MarshalsToJSON(frozenBundle, b2, "Thawed Bundle")

	// We have to set the username explicitly for the next test to pass, as a simple unmarshaling
	// of a bundle doesn't know user details (nor should it)
	require.DeepEqual(b, b2, "Thawed vs Created bundle")
}
