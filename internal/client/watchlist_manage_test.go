package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

// Locks the reverse-engineered new-watchlists mutation contract (method, path,
// body) so a future refactor can't silently drift.
func TestWatchlistMutationContract(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Write([]byte(`{"result":{"id":123,"name":"x","type":"USER_MADE","itemCount":0}}`))
	}))
	defer srv.Close()

	c := New(Config{
		HTTPClient:  srv.Client(),
		CertBaseURL: srv.URL,
		InfoBaseURL: srv.URL,
		Session:     &session.Session{Cookies: map[string]string{"SESSION": "x"}, Headers: map[string]string{"X-XSRF-TOKEN": "t"}},
	})

	// create group
	if _, err := c.CreateWatchlistGroup(context.Background(), "내폴더"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if gotMethod != "POST" || gotPath != "/api/v1/new-watchlists/groups" {
		t.Errorf("create routing: %s %s", gotMethod, gotPath)
	}
	var cbody map[string]string
	_ = json.Unmarshal([]byte(gotBody), &cbody)
	if cbody["name"] != "내폴더" {
		t.Errorf("create body: %s", gotBody)
	}

	// rename → PATCH /groups/{id}
	if err := c.RenameWatchlistGroup(context.Background(), 123, "새이름"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if gotMethod != "PATCH" || gotPath != "/api/v1/new-watchlists/groups/123" {
		t.Errorf("rename routing: %s %s", gotMethod, gotPath)
	}

	// delete → DELETE /groups/{id}
	if err := c.DeleteWatchlistGroup(context.Background(), 123); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if gotMethod != "DELETE" || gotPath != "/api/v1/new-watchlists/groups/123" {
		t.Errorf("delete routing: %s %s", gotMethod, gotPath)
	}
}

// XSRF header must ride along on mutations (via applySession session.Headers).
func TestWatchlistMutationSendsXSRF(t *testing.T) {
	var xsrf string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xsrf = r.Header.Get("X-XSRF-TOKEN")
		w.Write([]byte(`{"result":{"id":1}}`))
	}))
	defer srv.Close()
	c := New(Config{
		HTTPClient:  srv.Client(),
		CertBaseURL: srv.URL,
		Session:     &session.Session{Cookies: map[string]string{"SESSION": "x"}, Headers: map[string]string{"X-XSRF-TOKEN": "tok-123"}},
	})
	if _, err := c.CreateWatchlistGroup(context.Background(), "f"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if xsrf != "tok-123" {
		t.Errorf("expected XSRF header, got %q", xsrf)
	}
}
