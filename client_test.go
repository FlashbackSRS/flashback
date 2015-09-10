package flashback

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/html"
)

type ProcessATagData struct {
	name string
	html string
	rel  string
	href string
}

func TestNew(t *testing.T) {
	api := New("/oink")
	expected := "/oink/"
	if uri := api.URIs["index"]; uri != expected {
		t.Fatalf("New() set index URI to '%s', not '%s'", uri, expected)
	}
}

func TestProcessATag(t *testing.T) {
	tests := []ProcessATagData{
		//name      HTML                                            rel         href
		{"1", "<a href='foo.html' rel='foo'>Oink</a>", "foo", "foo.html"},
		{"2", "<a href=\"foo.html\" rel=\"foo\">Oink</a>", "foo", "foo.html"},
		{"3", "<a href='oink.html'></a>", "", ""},
		{"4", "<a href='a' rel='b' foo='bar' bar='baz'>", "b", "a"},
		{"5", "invalid html", "", ""},
	}
	for _, test := range tests {
		doc := fmt.Sprintf(test.html)
		z := html.NewTokenizer(strings.NewReader(doc))
		z.Next()
		tt := z.Token()

		rel, href := processATag(tt)
		if rel != test.rel {
			t.Fatalf("ProcessATag() test '%s' returned rel=%s, expected %s\n", test.name, rel, test.rel)
		}
		if href != test.href {
			t.Fatalf("ProcessATag() test '%s' returned href=%s, expected %s\n", test.name, href, test.href)
		}
	}
}

type ExtractLinksData struct {
	name  string
	html  string
	links map[string]string
}

func TestExtractLinks(t *testing.T) {
	tests := []ExtractLinksData{
		{
			"test1",
			"<a href='foo.html' rel='foo'>Oink</a>",
			map[string]string{
				"foo": "foo.html",
			},
		},
		{
			"test2",
			"<a href='foo.html'>Oink</a>",
			map[string]string{},
		},
		{
			"test3",
			"<title>Foo bar</title><a href='#trash' rel='trash'><a href='oink'></close>",
			map[string]string{
				"trash": "#trash",
			},
		},
	}
	for _, test := range tests {
		links := extractLinks(strings.NewReader(test.html))
		if !reflect.DeepEqual(test.links, links) {
			spew.Fprintf(os.Stderr, "extractLinks test '%s' failed.\nExpected:\n%v\n\nGot:\n%v\n",
				test.name, test.links, links)
			t.Fatal()
		}
	}
}

func TestGetURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
        <html>
        <head>
        </head>
        <body>
            <a href="/" rel="index"></a>
            <a href="pig.html" rel="login-facebook"></a>
        </body>
        </html>
        `)
	}))
	defer server.Close()
	api := New(server.URL)
	uri := api.GetURI("login-facebook")
	expected := server.URL + "/pig.html"
	if uri != expected {
		t.Fatalf("GetURI() returned '%s', expected '%s'", uri, expected)
	}
}
