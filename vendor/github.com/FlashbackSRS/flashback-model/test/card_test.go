package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback-model"
)

var frozenCard = []byte(`
{
	"_id": "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
	"type": "card",
	"created": "2016-07-31T15:08:24.730156517Z",
	"modified": "2016-07-31T15:08:24.730156517Z",
	"imported": "2016-08-02T15:08:24.730156517Z",
	"model": "theme-VGVzdCBUaGVtZQ/0",
	"due": "2017-01-01",
	"interval": 50
}
`)

func TestCard(t *testing.T) {
	require := require.New(t)
	b, _ := fb.NewBundle("bundle-krsxg5baij2w4zdmmu", "tui5ajfbabaeljnxt4om7fwmt4")
	c, err := fb.NewCard("theme-VGVzdCBUaGVtZQ", 0, "card-"+strings.TrimPrefix(b.ID, "bundle-")+".mViuXQThMLoh1G1Nlc4d_E8kR8o.0")
	require.Nil(err, "Error creating card: %s", err)

	c.Created = now
	c.Modified = now
	c.Imported = now.AddDate(0, 0, 2)
	due, _ := fb.ParseDue("2017-01-01")
	c.Due = due
	ivl, _ := fb.ParseInterval("50d")
	c.Interval = ivl
	require.MarshalsToJSON(frozenCard, c, "Created Card")

	c2 := &fb.Card{}
	err = json.Unmarshal(frozenCard, c2)
	require.Nil(err, "Error thawing card: %s", err)
	require.MarshalsToJSON(frozenCard, c2, "Thawed Card")

	require.DeepEqual(c, c2, "Thawed vs Created Cards")
}
