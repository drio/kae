package main

// NOTE: This way of testing (and the helper functions) come from
// https://benhoyt.com/writings/simple-lists/

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestServer(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()
	model, err := NewSQLModel(db)
	exitOnError(err)
	server, err := NewServer(model, log.Default())

	// Fetch homepage
	{
		recorder := serve(t, server, "GET", "/", nil)

		ensureCode(t, recorder, http.StatusOK)
		forms := parseForms(t, recorder.Body.String())
		ensureInt(t, len(forms), 1)
		ensureString(t, forms[0].Action, "/newtoken")
	}
}

// getText recursively assembles the text nodes of n into a string.
func getText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	text := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}
	return text
}

// ensureString asserts that got==want for strings.
func ensureString(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// ensureString asserts that got==want for ints.
func ensureInt(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

// getAttr returns the value of the named attribute, or "" if not found.
func getAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

type Form struct {
	Action string
	Inputs map[string]string
	Label  string
}

// parseForms parses the forms in an HTML document and returns the list of forms.
func parseForms(t *testing.T, htmlStr string) []Form {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("parsing HTML: %v", err)
	}

	var forms []Form
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			action := getAttr(n, "action")
			method := getAttr(n, "method")
			if method != "POST" {
				t.Fatalf("form %s method: got %s, want POST", action, method)
			}
			enctype := getAttr(n, "enctype")
			if enctype != "application/x-www-form-urlencoded" {
				t.Fatalf("form %s enctype: got %s, want application/x-www-form-urlencoded",
					action, method)
			}

			inputs := make(map[string]string)
			var label string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "input" {
					inputs[getAttr(c, "name")] = getAttr(c, "value")
				}
				if c.Type == html.ElementNode && c.Data == "label" {
					label = getText(c.FirstChild)
				}
			}

			forms = append(forms, Form{
				Action: action,
				Inputs: inputs,
				Label:  label,
			})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return forms
}

// ensureCode asserts that the HTTP status code is correct.
func ensureCode(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("got code %d, want %d, response body:\n%s",
			recorder.Code, expected, recorder.Body.String())
	}
}

// serve records a single HTTP request and returns the response recorder.
func serve(t *testing.T, server *Server, method, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	r, err := http.NewRequest(method, "http://localhost"+path, body)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	if form != nil {
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, r)
	return recorder
}
