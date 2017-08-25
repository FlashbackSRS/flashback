package oauth2

import (
	"fmt"
	"net/url"
)

const facebookAuthURL = "https://www.facebook.com/v2.9/dialog/oauth"

// FacebookURL returns the Facebook auth URL to used, based on the configuration.
func FacebookURL(clientID, appURL string) string {
	redirParams := url.Values{}
	redirParams.Add("provider", "facebook")
	redir, err := url.Parse(appURL)
	if err != nil {
		panic("Invalid flashback_app URL: " + err.Error())
	}
	redir.RawQuery = redirParams.Encode()

	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("redirect_uri", redir.String())
	params.Add("response_type", "token")
	return fmt.Sprintf("%s?%s", facebookAuthURL, params.Encode())
}
