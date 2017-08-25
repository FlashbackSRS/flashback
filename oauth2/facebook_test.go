package oauth2

import (
	"fmt"
	"testing"
)

func TestFacebookURL(t *testing.T) {
	url := FacebookURL("12345", "https://foo.com/app")
	expected := fmt.Sprintf("%s?%s", facebookAuthURL, `client_id=12345&redirect_uri=https%3A%2F%2Ffoo.com%2Fapp%3Fprovider%3Dfacebook&response_type=token`)
	if url != expected {
		t.Errorf("Got  %s\nWant %s\n", url, expected)
	}
	t.Run("InvalidURL", func(t *testing.T) {
		err := func() (err string) {
			defer func() {
				if r := recover(); r != nil {
					err = r.(string)
				}
			}()
			_ = FacebookURL("12345", "https://foo.com/app%xx")
			return
		}()
		if err != `Invalid flashback_app URL: parse https://foo.com/app%xx: invalid URL escape "%xx"` {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}
