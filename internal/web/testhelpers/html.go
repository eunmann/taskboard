// Package testhelpers provides HTML testing utilities using goquery.
package testhelpers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// ParseHTML parses the response body into a goquery document.
func ParseHTML(t *testing.T, w *httptest.ResponseRecorder) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(w.Body.String()))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}

	return doc
}

// AssertStatus checks the response status code.
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if w.Code != expected {
		t.Errorf("status = %d, want %d", w.Code, expected)
	}
}

// AssertElementExists checks that a CSS selector matches at least one element.
func AssertElementExists(t *testing.T, doc *goquery.Document, selector string) {
	t.Helper()

	if doc.Find(selector).Length() == 0 {
		t.Errorf("element not found: %s", selector)
	}
}

// AssertElementCount checks the number of elements matching a selector.
func AssertElementCount(t *testing.T, doc *goquery.Document, selector string, count int) {
	t.Helper()

	actual := doc.Find(selector).Length()
	if actual != count {
		t.Errorf("element count for %q = %d, want %d", selector, actual, count)
	}
}

// AssertElementText checks the text content of the first matching element.
func AssertElementText(t *testing.T, doc *goquery.Document, selector, text string) {
	t.Helper()

	actual := strings.TrimSpace(doc.Find(selector).First().Text())
	if actual != text {
		t.Errorf("text for %q = %q, want %q", selector, actual, text)
	}
}

// AssertTitle checks the page title.
func AssertTitle(t *testing.T, doc *goquery.Document, expected string) {
	t.Helper()

	actual := strings.TrimSpace(doc.Find("title").Text())
	if actual != expected {
		t.Errorf("title = %q, want %q", actual, expected)
	}
}

// AssertHTMXAttr checks an HTMX attribute on the first matching element.
func AssertHTMXAttr(t *testing.T, doc *goquery.Document, selector, attr, value string) {
	t.Helper()

	el := doc.Find(selector).First()
	actual, exists := el.Attr(attr)

	if !exists {
		t.Errorf("attribute %q not found on %q", attr, selector)

		return
	}

	if actual != value {
		t.Errorf("attribute %q on %q = %q, want %q", attr, selector, actual, value)
	}
}

// AssertTableRowCount checks the number of tbody rows in a table.
func AssertTableRowCount(t *testing.T, doc *goquery.Document, tableSelector string, count int) {
	t.Helper()

	actual := doc.Find(tableSelector + " tbody tr").Length()
	if actual != count {
		t.Errorf("table row count for %q = %d, want %d", tableSelector, actual, count)
	}
}
