package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

var database = make(map[string]string)

type shortenreq struct {
	URLs []string `json:"urls"`
}

type shortenres struct {
	Results []URLPair `json:"results"`
}

type URLPair struct {
	Original string `json:"original"`
	Short    string `json:"short"`
}

func shortstring() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shortenreq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.URLs) == 0 {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	var response shortenres

	for _, originalURL := range req.URLs {
		id := shortstring()
		database[id] = originalURL

		response.Results = append(response.Results, URLPair{
			Original: originalURL,
			Short:    fmt.Sprintf("http://localhost:8080/%s", id),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	longURL, found := database[id]
	if !found {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Bulk API Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
