package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OWNER/PROJECT_NAME/internal/web"
	"github.com/OWNER/PROJECT_NAME/internal/web/testhelpers"
)

func TestAboutPage(t *testing.T) {
	renderer, err := web.NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	handlers := web.NewHandlers(renderer, nil)

	req := httptest.NewRequest(http.MethodGet, "/about", http.NoBody)
	w := httptest.NewRecorder()

	handlers.About(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)
	testhelpers.AssertTitle(t, doc, "About - App")
	testhelpers.AssertElementExists(t, doc, "h1")
}
