// +build js

package loginhandler

import (
	"fmt"
	"testing"

	"github.com/FlashbackSRS/flashback/config"
)

func TestFacebookURL(t *testing.T) {
	c := config.New(map[string]string{"facebook_client_id": "12345", "flashback_app": "https://foo.com/app"})
	url := facebookURL(c)
	expected := fmt.Sprintf("%s?%s", facebookAuthURL, `client_id=12345&redirect_uri=https%3A%2F%2Ffoo.com%2Fapp`)
	if url != expected {
		t.Errorf("Got  %s\nWant %s\n", url, expected)
	}
}
