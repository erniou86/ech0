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

func openTestDB() (*sql.DB, error) {
	return sql.Open("sqlite3", os.Getenv("DB_PATH"))
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	name := "test-" + strings.ReplaceAll(t.Name(), "/", "-") + ".db"
	if err := os.Setenv("DB_PATH", name); err != nil {
		t.Fatal(err)
	}

	var err error
	db, err = openTestDB()
	if err != nil {
		t.Fatal(err)
	}
	if err := initDB(); err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(route))
	t.Cleanup(func() {
		ts.Close()
		db.Close()
		_ = os.Remove(name)
	})
	return ts
}

func TestHealth(t *testing.T) {
	ts := setupTestServer(t)
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRegisterUserAndCreateRoom(t *testing.T) {
	ts := setupTestServer(t)

	resp, err := http.Post(ts.URL+"/api/users/register", "application/json", bytes.NewBufferString(`{"display_name":"alice"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatal(err)
	}
	if user.DisplayName != "alice" {
		t.Fatalf("unexpected user name: %s", user.DisplayName)
	}

	resp2, err := http.Post(ts.URL+"/api/rooms", "application/json", bytes.NewBufferString(`{"name":"general","description":"welcome"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp2.StatusCode)
	}

	var room Room
	if err := json.NewDecoder(resp2.Body).Decode(&room); err != nil {
		t.Fatal(err)
	}
	if room.Name != "general" {
		t.Fatalf("unexpected room name: %s", room.Name)
	}

	resp3, err := http.Post(ts.URL+"/api/rooms/"+strconv.Itoa(room.ID)+"/messages", "application/json", bytes.NewBufferString(`{"user_id":`+strconv.Itoa(user.ID)+`,"content":"hello everyone"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp3.StatusCode)
	}

	resp4, err := http.Get(ts.URL + "/api/rooms/" + strconv.Itoa(room.ID) + "/messages")
	if err != nil {
		t.Fatal(err)
	}
	defer resp4.Body.Close()
	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp4.StatusCode)
	}

	var messages []Message
	if err := json.NewDecoder(resp4.Body).Decode(&messages); err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Content != "hello everyone" {
		t.Fatalf("unexpected messages: %+v", messages)
	}
}

func TestListRooms(t *testing.T) {
	ts := setupTestServer(t)
	_, err := http.Post(ts.URL+"/api/rooms", "application/json", bytes.NewBufferString(`{"name":"tech","description":"dev room"}`))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(ts.URL + "/api/rooms")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var rooms []Room
	if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 1 || rooms[0].Name != "tech" {
		t.Fatalf("unexpected room list: %+v", rooms)
	}
}

func route(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		indexHandler(w, r)
	case r.URL.Path == "/health":
		healthHandler(w, r)
	case r.URL.Path == "/api/users/register":
		registerUserHandler(w, r)
	case r.URL.Path == "/api/rooms":
		roomsHandler(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/rooms/"):
		roomMessagesHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}
