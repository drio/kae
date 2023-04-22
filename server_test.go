package main

// NOTE: This way of testing (and the helper functions) come from
// https://benhoyt.com/writings/simple-lists/

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
	if err != nil {
		t.Fatalf("Error creating server")
	}

	// Fetch homepage
	{
		recorder := serve(t, server, "GET", "/", nil)

		ensureCode(t, recorder, http.StatusOK)
		forms := parseForms(t, recorder.Body.String())
		ensureInt(t, len(forms), 1)
		ensureString(t, forms[0].Action, "/newtoken")
	}

	// Create a token
	{
		form := url.Values{}
		form.Set("name", "token one")
		form.Set("interval", "5")
		form.Set("description", "the first token")
		recorder := serve(t, server, "POST", "/newtoken", form)

		ensureCode(t, recorder, http.StatusFound)
		location := recorder.Result().Header.Get("Location")
		ensureString(t, location, "/")
	}

	// Create another token
	{
		time.Sleep(time.Millisecond) // wait at least 1ms to ensure time_created is newer
		form := url.Values{}
		form.Set("name", "token two")
		form.Set("interval", "2")
		form.Set("description", "the second token")
		recorder := serve(t, server, "POST", "/newtoken", form)
		ensureCode(t, recorder, http.StatusFound)
		location := recorder.Result().Header.Get("Location")
		ensureString(t, location, "/")
	}

	// Fetch homepage again (should have two tokens)
	{
		recorder := serve(t, server, "GET", "/", nil)

		links := parseLinks(t, recorder.Body.String())
		ensureInt(t, len(links), 4) // 2 tokens, each has a delete and enable
		ensureString(t, links[0].Href, "/delete/2")
		ensureString(t, links[0].Text, "delete")
		ensureString(t, links[1].Href, "/enable/2")
		ensureString(t, links[1].Text, "enable")
		ensureString(t, links[2].Href, "/delete/1")
		ensureString(t, links[2].Text, "delete")
		ensureString(t, links[3].Href, "/enable/1")
		ensureString(t, links[3].Text, "enable")
	}

	// Enable a token
	{
		recorder := serve(t, server, "GET", "/enable/2", nil)
		location := recorder.Result().Header.Get("Location")
		ensureString(t, location, "/")
		ensureCode(t, recorder, http.StatusFound)
	}

	// Now check in the UI to make sure the second token can be disabled
	{
		recorder := serve(t, server, "GET", "/", nil)
		links := parseLinks(t, recorder.Body.String())
		ensureInt(t, len(links), 4)
		ensureString(t, links[0].Href, "/delete/2")
		ensureString(t, links[0].Text, "delete")
		ensureString(t, links[1].Href, "/disable/2")
		ensureString(t, links[1].Text, "disable")
	}

	// Delete a token
	{
		recorder := serve(t, server, "GET", "/delete/2", nil)
		location := recorder.Result().Header.Get("Location")
		ensureString(t, location, "/")
		ensureCode(t, recorder, http.StatusFound)
	}

	// Check that we only have a token now
	{
		recorder := serve(t, server, "GET", "/", nil)
		links := parseLinks(t, recorder.Body.String())
		ensureInt(t, len(links), 2)
		ensureString(t, links[0].Href, "/delete/1")
		ensureString(t, links[0].Text, "delete")
		ensureString(t, links[1].Href, "/enable/1")
		ensureString(t, links[1].Text, "enable")
	}

	// Get the token value and send a heartbeat
	{
		recorder := serve(t, server, "GET", "/", nil)
		divs := parseGeneric(t, recorder.Body.String(), "div", "token-value")
		ensureInt(t, len(divs), 1)
		_ = serve(t, server, "GET", "/hb/"+divs[0].Text, nil)
	}

	// Enable the token
	{
		recorder := serve(t, server, "GET", "/enable/1", nil)
		location := recorder.Result().Header.Get("Location")
		ensureString(t, location, "/")
		ensureCode(t, recorder, http.StatusFound)
	}

	// The UI should tell us that the state of the token is fire because
	// we sent a heartbeat but we didn't run the background job
	{
		recorder := serve(t, server, "GET", "/", nil)
		divs := parseGeneric(t, recorder.Body.String(), "span", "emoji")
		fmt.Printf("%s\n", recorder.Body.String())
		ensureInt(t, len(divs), 1)
		ensureString(t, divs[0].Text, "🔥")
	}

	// Run the background job
	{
		server.runBackgroundJob(0)
	}

	// The UI should tell us now that the token is not in fire state
	{
		recorder := serve(t, server, "GET", "/", nil)
		divs := parseGeneric(t, recorder.Body.String(), "span", "emoji")
		fmt.Printf("%s\n", recorder.Body.String())
		ensureInt(t, len(divs), 1)
		ensureString(t, divs[0].Text, "🟢")
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

// ensureCode asserts that the HTTP status code is correct.
func ensureCode(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("got code %d, want %d, response body:\n%s",
			recorder.Code, expected, recorder.Body.String())
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

type Link struct {
	Href string
	Text string
}

// parseLinks parses the links in an HTML document and returns the list of links.
func parseLinks(t *testing.T, htmlStr string) []Link {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("parsing HTML: %v", err)
	}

	var links []Link
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			links = append(links, Link{
				Href: getAttr(n, "href"),
				Text: getText(n),
			})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return links
}

type Div struct {
	Class string
	Text  string
}

// parseDivs parses the div in an HTML document and returns the list of divs.
func parseGeneric(t *testing.T, htmlStr string, nodeType string, className string) []Div {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("parsing HTML: %v", err)
	}

	var divs []Div
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Data == nodeType && len(n.Attr) > 0 {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == className {
					divs = append(divs, Div{
						Text: getText(n),
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return divs
}
