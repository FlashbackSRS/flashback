package model

import (
	"context"
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
)

func TestFetchAttachment(t *testing.T) {
	tests := []struct {
		name     string
		repo     *Repo
		cardID   string
		filename string
		expected *fb.Attachment
		err      string
	}{
		{
			name:   "not logged in",
			repo:   &Repo{},
			cardID: "card-foo.bar.0",
			err:    "not logged in",
		},
		// {
		// 	name: "user db not found",
		// 	repo: &Repo{
		// 		user:  "bob",
		// 		local: testClient(t),
		// 	},
		// 	cardID: "card-foo.bar.0",
		// 	err:    "database does not exist",
		// },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.repo.FetchAttachment(context.Background(), test.cardID, test.filename)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpectd error: %s", errMsg)
			}
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
