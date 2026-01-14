package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

// maps short url to long one
var database = make(map[string]string)

type shortenreq struct {
	URLs []string
}
type shortenres struct {
	Results []URLPair
}
type URLPair struct {
	Original string
	Short    string
}

func shortstring() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		longURL, found := database[id]
		if !found {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, longURL, http.StatusFound)
	})

	fmt.Println("Bulk API Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
