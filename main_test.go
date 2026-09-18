package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func openDB() (*sql.DB, error) {
	return sql.Open("sqlite3", os.Getenv("DB_PATH"))
}

func setup(t *testing.T) *httptest.Server {
	t.Helper()
	os.Setenv("DB_PATH", "./test-"+strings.ReplaceAll(t.Name(), "/", "-")+".db")
	var err error
	db, err = openDB()
	if err != nil {
		t.Fatal(err)
	}
	initDB()
	ts := httptest.NewServer(http.HandlerFunc(route))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
	})
	return ts
}

func TestCreateAndListNotes(t *testing.T) {
	ts := setup(t)
	body := `{"content":"hello ech0","tags":"go,test"}`
	resp, err := http.Post(ts.URL+"/api/notes", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var n Note
	json.NewDecoder(resp.Body).Decode(&n)
	if n.Content != "hello ech0" {
		t.Fatalf("unexpected content: %s", n.Content)
	}

	resp2, _ := http.Get(ts.URL + "/api/notes")
	defer resp2.Body.Close()
	var notes []Note
	json.NewDecoder(resp2.Body).Decode(&notes)
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
}

func TestSearchNotes(t *testing.T) {
	ts := setup(t)
	for _, c := range []string{"apple pie", "banana bread", "cherry"} {
		http.Post(ts.URL+"/api/notes", "application/json", bytes.NewBufferString(`{"content":"`+c+`"}`))
	}
	resp, _ := http.Get(ts.URL + "/api/notes?q=banana")
	defer resp.Body.Close()
	var notes []Note
	json.NewDecoder(resp.Body).Decode(&notes)
	if len(notes) != 1 || notes[0].Content != "banana bread" {
		t.Fatalf("search failed: %+v", notes)
	}
}

func TestDeleteNote(t *testing.T) {
	ts := setup(t)
	http.Post(ts.URL+"/api/notes", "application/json", bytes.NewBufferString(`{"content":"to delete"}`))
	resp, _ := http.Get(ts.URL + "/api/notes")
	var notes []Note
	json.NewDecoder(resp.Body).Decode(&notes)
	resp.Body.Close()

	req, _ := http.NewRequest("DELETE", ts.URL+"/api/notes/"+strconv.Itoa(notes[0].ID), nil)
	resp2, _ := http.DefaultClient.Do(req)
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("delete failed: %d", resp2.StatusCode)
	}
}

func TestRSS(t *testing.T) {
	ts := setup(t)
	http.Post(ts.URL+"/api/notes", "application/json", bytes.NewBufferString(`{"content":"rss item"}`))
	resp, _ := http.Get(ts.URL + "/api/rss")
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if !strings.Contains(buf.String(), "<rss") || !strings.Contains(buf.String(), "rss item") {
		t.Fatalf("rss content wrong: %s", buf.String())
	}
}

func TestHealth(t *testing.T) {
	ts := setup(t)
	resp, _ := http.Get(ts.URL + "/health")
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("health check failed")
	}
}

func route(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		indexHandler(w, r)
	case r.URL.Path == "/api/notes":
		notesHandler(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/notes/"):
		noteHandler(w, r)
	case r.URL.Path == "/api/rss":
		rssHandler(w, r)
	case r.URL.Path == "/health":
		w.Write([]byte("ok"))
	default:
		http.NotFound(w, r)
	}
}
