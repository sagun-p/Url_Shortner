package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetDatabase() {
	database = make(map[string]string)
}

func TestShortenHandler_Success(t *testing.T) {
	resetDatabase()

	body := shortenreq{
		URLs: []string{
			"https://example.com",
			"https://golang.org",
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()
	shortenHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var result shortenres
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	for _, pair := range result.Results {
		if pair.Original == "" || pair.Short == "" {
			t.Fatal("result fields should not be empty")
		}
	}
}

func TestShortenHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatal("expected 405 Method Not Allowed")
	}
}

func TestShortenHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		bytes.NewBufferString("invalid-json"),
	)
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400 Bad Request")
	}
}

func TestShortenHandler_EmptyURLList(t *testing.T) {
	resetDatabase()

	body := shortenreq{URLs: []string{}}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400 Bad Request for empty url list")
	}
}

func TestShortenHandler_MissingURLsField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	shortenHandler(w, req)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400 Bad Request")
	}
}

func TestShortenHandler_InvalidURL(t *testing.T) {
	resetDatabase()

	body := shortenreq{
		URLs: []string{"not-a-url"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400 Bad Request")
	}
}

func TestShortenHandler_UnsupportedScheme(t *testing.T) {
	resetDatabase()

	body := shortenreq{
		URLs: []string{"ftp://example.com"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400 Bad Request")
	}
}

func TestShortenHandler_DuplicateURLs(t *testing.T) {
	resetDatabase()

	body := shortenreq{
		URLs: []string{
			"https://example.com",
			"https://example.com",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result shortenres
	_ = json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Results) != 2 {
		t.Fatal("expected 2 results")
	}
}

func TestShortenHandler_VeryLongURL(t *testing.T) {
	resetDatabase()

	longURL := "https://example.com/" + string(bytes.Repeat([]byte("a"), 5000))
	body := shortenreq{URLs: []string{longURL}}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("expected long url to be handled")
	}
}

func TestShortenHandler_NoCollision(t *testing.T) {
	resetDatabase()

	body := shortenreq{
		URLs: []string{
			"https://a.com",
			"https://b.com",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	shortenHandler(w, req)

	var result shortenres
	_ = json.NewDecoder(w.Result().Body).Decode(&result)

	if result.Results[0].Short == result.Results[1].Short {
		t.Fatal("collision detected")
	}
}

func TestRedirectHandler_Success(t *testing.T) {
	resetDatabase()

	database["abc123"] = "https://example.com"

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	redirectHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 Found, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Location") != "https://example.com" {
		t.Fatal("unexpected redirect location")
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	resetDatabase()

	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	w := httptest.NewRecorder()

	redirectHandler(w, req)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatal("expected 404 Not Found")
	}
}
