package handlers

import (
	"book-api/models"
	"book-api/repository"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SearchHandler struct {
	repo repository.BookRepository
}

func NewSearchHandler(repo repository.BookRepository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

func (h *SearchHandler) ExecuteBookSearch(w http.ResponseWriter, r *http.Request) {
	// Start with full request logging
	log.Printf("\n=== SEARCH REQUEST STARTED ===")
	log.Printf("Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.String())

	// Validate and log query parameter
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	log.Printf("Raw query parameter: %q", query)
	log.Printf("Lowercase query: %q", strings.ToLower(query))

	if err := validateSearchQuery(query); err != nil {
		log.Printf("Invalid query: %v", err)
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Database operation with timing and full data dump
	dbStart := time.Now()
	books, err := h.repo.GetAllBooks()
	if err != nil {
		log.Printf("CRITICAL DATABASE ERROR: %v", err)
		sendJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	log.Printf("Loaded %d books in %v", len(books), time.Since(dbStart))

	// TEMPORARY: Full dump of all books for debugging
	log.Printf("\n=== COMPLETE BOOKS DUMP ===")
	for i, book := range books {
		log.Printf("Book %d:", i+1)
		log.Printf("  ID: %s", book.BookID)
		log.Printf("  Title: %q", book.Title)
		log.Printf("  Description: %q", book.Description)
		log.Printf("  Genre: %q", book.Genre)
		log.Printf("  AuthorID: %q", book.AuthorID)
		log.Printf("  PublisherID: %q", book.PublisherID)
	}
	log.Printf("=== END BOOKS DUMP ===\n")

	// Perform search with detailed matching logs
	searchStart := time.Now()
	matchedBooks := h.searchBooks(books, query)
	searchDuration := time.Since(searchStart)

	log.Printf("\nSearch completed in %v", searchDuration)
	log.Printf("Query: %q", query)
	log.Printf("Books scanned: %d", len(books))
	log.Printf("Matches found: %d", len(matchedBooks))

	if len(matchedBooks) == 0 {
		log.Printf("NO MATCHES FOUND for query %q in any books", query)
		sendJSONError(w, http.StatusNotFound, "No books found matching '"+query+"'")
		return
	}

	// Log match details
	log.Printf("\n=== MATCHES FOUND ===")
	for i, book := range matchedBooks {
		if i >= 5 { // Limit to first 5 matches
			break
		}
		log.Printf("Match %d:", i+1)
		log.Printf("  ID: %s", book.BookID)
		log.Printf("  Title: %q", book.Title)
		log.Printf("  Description: %q", book.Description)
		log.Printf("  Genre: %q", book.Genre)
	}

	sendJSONResponse(w, http.StatusOK, matchedBooks)
	log.Printf("\n=== REQUEST COMPLETED IN %v ===\n", time.Since(dbStart))
}

func (h *SearchHandler) searchBooks(books []*models.Book, query string) []*models.Book {
	if len(books) == 0 {
		log.Printf("WARNING: Empty books list provided to search")
		return nil
	}

	log.Printf("Starting search with strategy selection...")
	if len(books) < 20 {
		log.Printf("Using simple sequential search for %d books", len(books))
		return h.simpleSearch(books, query)
	}

	log.Printf("Using concurrent search for %d books", len(books))
	return h.concurrentSearch(books, query)
}

func (h *SearchHandler) simpleSearch(books []*models.Book, query string) []*models.Book {
	var matches []*models.Book
	lowerQuery := strings.ToLower(query)

	for _, book := range books {
		if containsSearchTerm(book, lowerQuery) {
			matches = append(matches, book)
		}
	}
	return matches
}

func (h *SearchHandler) concurrentSearch(books []*models.Book, query string) []*models.Book {
	chunkSize := calculateChunkSize(len(books))
	chunks := chunkBooks(books, chunkSize)
	results := make(chan []*models.Book)
	var wg sync.WaitGroup
	lowerQuery := strings.ToLower(query)

	for _, chunk := range chunks {
		wg.Add(1)
		go func(c []*models.Book) {
			defer wg.Done()
			var chunkMatches []*models.Book
			for _, book := range c {
				if containsSearchTerm(book, lowerQuery) {
					chunkMatches = append(chunkMatches, book)
				}
			}
			results <- chunkMatches
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var matchedBooks []*models.Book
	for matches := range results {
		matchedBooks = append(matchedBooks, matches...)
	}

	return matchedBooks
}

func containsSearchTerm(book *models.Book, query string) bool {
	lowerQuery := strings.ToLower(query)

	fields := []struct {
		name  string
		value string
	}{
		{"Title", book.Title},
		{"Description", book.Description},
		{"Genre", book.Genre},
	}

	for _, field := range fields {
		lowerValue := strings.ToLower(field.value)
		if strings.Contains(lowerValue, lowerQuery) {
			log.Printf("MATCH FOUND in Book %q:", book.BookID)
			log.Printf("  Field: %s", field.name)
			log.Printf("  Value: %q", field.value)
			log.Printf("  Contains: %q", query)
			return true
		}
		log.Printf("No match in Book %q field %q (%q doesn't contain %q)",
			book.BookID, field.name, field.value, query)
	}

	return false
}

func calculateChunkSize(totalBooks int) int {
	const (
		minChunkSize = 20
		maxChunkSize = 50
		idealWorkers = 4
	)

	chunkSize := totalBooks / idealWorkers
	switch {
	case chunkSize < minChunkSize:
		return minChunkSize
	case chunkSize > maxChunkSize:
		return maxChunkSize
	default:
		return chunkSize
	}
}

func chunkBooks(books []*models.Book, chunkSize int) [][]*models.Book {
	var chunks [][]*models.Book
	for i := 0; i < len(books); i += chunkSize {
		end := i + chunkSize
		if end > len(books) {
			end = len(books)
		}
		chunks = append(chunks, books[i:end])
	}
	return chunks
}

func validateSearchQuery(query string) error {
	if query == "" {
		return &SearchError{"Search query cannot be empty"}
	}
	if len(query) < 2 {
		return &SearchError{"Search query must be at least 2 characters long"}
	}
	if len(query) > 100 {
		return &SearchError{"Search query too long (max 100 characters)"}
	}
	return nil
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func sendJSONError(w http.ResponseWriter, statusCode int, message string) {
	response := struct {
		Error string `json:"error"`
	}{message}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type SearchError struct {
	Message string
}

func (e *SearchError) Error() string {
	return e.Message
}
