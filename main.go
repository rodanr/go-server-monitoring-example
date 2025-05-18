package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	startTime time.Time
	database  *DB
)

var (
	apiCallCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_call_count",
			Help: "Tracks the number of API calls made to each endpoint.",
		},
		[]string{"endpoint", "method"},
	)

	uptimeGauge = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "uptime_seconds",
			Help: "The uptime of the application in seconds.",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)
)

func handleBooks(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/books")
	path = strings.Trim(path, "/")

	apiCallCounter.WithLabelValues("/books", r.Method).Inc()

	switch r.Method {
	case http.MethodGet:
		if path == "" {
			books := database.GetBooks()
			jsonResponse(w, books, http.StatusOK)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}
		book, err := database.GetBook(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		jsonResponse(w, book, http.StatusOK)
		return

	case http.MethodPost:
		var book Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		newBook := database.AddBook(book.Name, book.Author)
		jsonResponse(w, newBook, http.StatusCreated)

	case http.MethodPut:
		if path == "" {
			http.Error(w, "Book ID is required", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}
		var book Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		updatedBook, err := database.UpdateBook(id, book.Name, book.Author)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonResponse(w, updatedBook, http.StatusOK)

	case http.MethodDelete:
		if path == "" {
			http.Error(w, "Book ID is required", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}
		if err := database.RemoveBook(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func jsonResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	data := map[string]any{
		"status": "ok",
		"uptime": fmt.Sprintf("%s", uptime.Round(time.Second)),
	}

	jsonResponse(w, data, http.StatusOK)
}

func main() {
	prometheus.MustRegister(apiCallCounter)
	prometheus.MustRegister(uptimeGauge)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/books", handleBooks)
	http.HandleFunc("/books/", handleBooks)

	log.Println("Starting server on :2112")
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func init() {
	startTime = time.Now()

	database = NewDB()
	database.AddBook("Go Programming", "John Doe")
	database.AddBook("Concurrency in Go", "Jane Smith")
}
