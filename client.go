package flashback

import (
	"io"
	"regexp"
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/gopherjs/gopherjs/js"
)

type FlashbackClient struct {
	URIs map[string]string
}

func findRelURI() string {
	links := js.Global.Call("getElementByTagName", "link")
	for i := 0; i < links.Length(); i++ {
		fmt.Printf("%d. %s\n", i, links.Index(i))
	}
	return ""
}

var defaultFbClient *FlashbackClient
func NewWithURI(uri string) *FlashbackClient {
	return &FlashbackClient{
		map[string]string{
			"index": strings.TrimRight(uri, "/") + "/",
		},
	}
}

func New() *FlashbackClient {
	if defaultFbClient == nil {
		uri := findRelURI()
		defaultFbClient = NewWithURI(uri)
	}
	return defaultFbClient
}

var recognizedRels = []string{
	"index",
	"login",
	"login-facebook",
	"login-google",
}

func isRecognizedRel(rel string) bool {
	for _, r := range recognizedRels {
		if r == rel {
			return true
		}
	}
	return false
}

func (c *FlashbackClient) GetURI(rel string) string {
	if !isRecognizedRel(rel) {
		Console("Request for unrecognized rel '%s'", rel)
		return ""
	}
	if href, ok := c.URIs[rel]; ok {
		return href
	}
	if rel == "index" {
		panic("fbclient index not set !!?!?!")
	}
	c.ReadRels("index")
	return c.URIs[rel]
}

func (c *FlashbackClient) ReadRels(url string) {
	doc, err := c.Get(c.GetURI("index"))
	if err != nil {
		Console("Error reading index: %s", err)
		return
	}
	c.ExtractLinks(strings.NewReader(doc))
}

var absoluteURI = regexp.MustCompile("^https?://")

func (c *FlashbackClient) SetURI(rel, href string) {
	if isRecognizedRel(rel) {
		if !absoluteURI.MatchString(href) {
			href = c.GetURI("index") + strings.TrimLeft(href, "/")
		}
		c.URIs[rel] = href
	}
}

func (c *FlashbackClient) ExtractLinks(doc io.Reader) []string {
	links := extractLinks(doc)
	rels := make([]string, 0, len(links))
	for rel, href := range links {
		c.SetURI(rel, href)
		rels = append(rels, rel)
	}
	return rels
}

func extractLinks(doc io.Reader) map[string]string {
	z := html.NewTokenizer(doc)
	links := make(map[string]string)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			{
				// End of (parsable) document
				return links
			}
		case tt == html.StartTagToken:
			{
				if rel, href := processATag(z.Token()); len(rel) > 0 {
					links[rel] = href
				}
			}
		}
	}
}

func processATag(t html.Token) (string, string) {
	var rel, href string
	for _, a := range t.Attr {
		switch a.Key {
		case "rel":
			{
				rel = a.Val
			}
		case "href":
			{
				href = a.Val
			}
		}
	}
	if len(rel) > 0 && len(href) > 0 {
		return rel, href
	}
	return "", ""
}

func (c *FlashbackClient) GetLoginProviders() map[string]string {
	links := make(map[string]string)
	loginPage, err := c.Get(c.GetURI("login"))
	if err == nil {
		rels := c.ExtractLinks(strings.NewReader(loginPage))
		for _, rel := range rels {
			if pName := strings.TrimPrefix(rel, "login-"); pName != rel {
				links[pName] = c.GetURI(rel)
			}
		}
	}
	return links
}
