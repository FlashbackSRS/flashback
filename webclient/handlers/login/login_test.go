// +build js

package loginhandler

import "testing"

func TestExtractAuthToken(t *testing.T) {
	type eatTest struct {
		name     string
		url      string
		provider string
		token    string
		err      string
	}
	tests := []eatTest{
		{
			name: "InvalidURL",
			url:  "http://foo.com/%xxfoobar",
			err:  `parse http://foo.com/%xxfoobar: invalid URL escape "%xx"`,
		},
		{
			name: "UnknownProvider",
			url:  "http://foo.com/?provider=chicken",
			err:  "Unknown provider 'chicken'",
		},
		{
			name:     "Valid",
			url:      "https://flashback.ddns.net:4001/app/?provider=facebook#access_token=EAAMU5gISZBXoBAMXeg6zZCUNu7HhIiVgp0vk34XAFjiF3AYOm9Il3JkaeKAilHYfuuZArq5ZCypoGu67ZCgMT8SDqmj2Spiqy9qZB6gfIPGw9XGSYUr83DGZAthfEe7mv5I3hlNDqE622mAdZBWi7zQ3EAJoYSknVRVapvq13zd8bVCo7RrbHbyQqs3t1aRZBwj0ZD&expires_in=4839",
			provider: "facebook",
			token:    "EAAMU5gISZBXoBAMXeg6zZCUNu7HhIiVgp0vk34XAFjiF3AYOm9Il3JkaeKAilHYfuuZArq5ZCypoGu67ZCgMT8SDqmj2Spiqy9qZB6gfIPGw9XGSYUr83DGZAthfEe7mv5I3hlNDqE622mAdZBWi7zQ3EAJoYSknVRVapvq13zd8bVCo7RrbHbyQqs3t1aRZBwj0ZD",
		},
		{
			name: "NoProvider",
			url:  "https://foo.com/",
			err:  "no provider",
		},
	}
	for _, test := range tests {
		func(test eatTest) {
			t.Run(test.name, func(t *testing.T) {
				provider, token, err := extractAuthToken(test.url)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.err {
					t.Errorf("Unexpected error: %s", msg)
					return
				}
				if provider != test.provider {
					t.Errorf("Unexpected provider: %s", provider)
				}
				if token != test.token {
					t.Errorf("Unexpected token: %s", token)
				}
			})
		}(test)
	}
}
