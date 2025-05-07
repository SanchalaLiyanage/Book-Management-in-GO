package main

import (
	"book-api/handlers"
	"book-api/repository"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize storage with file-based persistence
	dataFilePath := getEnv("DATA_FILE_PATH", "data/books.json")
	store := repository.NewFileStore(dataFilePath)
	bookRepo := repository.NewBookRepository(store)

	bookHandler := handlers.NewBookHandler(bookRepo)
	searchHandler := handlers.NewSearchHandler(bookRepo)

	router := configureRouter(bookHandler, searchHandler)

	startServer(router)
}

func configureRouter(bookHandler *handlers.BookHandler, searchHandler *handlers.SearchHandler) *mux.Router {
	r := mux.NewRouter()

	r.Use(jsonContentTypeMiddleware)
	r.Use(requestLoggingMiddleware)
	r.Use(corsMiddleware)

	r.HandleFunc("/books", bookHandler.GetBooks).Methods("GET")
	r.HandleFunc("/books", bookHandler.CreateBook).Methods("POST")
	r.HandleFunc("/books/{id}", bookHandler.GetBook).Methods("GET")
	r.HandleFunc("/books/{id}", bookHandler.UpdateBook).Methods("PUT")
	r.HandleFunc("/books/{id}", bookHandler.DeleteBook).Methods("DELETE")

	r.HandleFunc("/books/search", searchHandler.ExecuteBookSearch).Methods("GET")
	// r.HandleFunc("/books/search/advanced", searchHandler.AdvancedBookSearch).Methods("GET")

	return r
}

func startServer(router *mux.Router) {
	port := getServerPort()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("Starting server on :%s...", port)
		log.Printf("Available routes:")
		if err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, _ := route.GetMethods()
			path, _ := route.GetPathTemplate()
			log.Printf("%-6s %s", methods, path)
			return nil
		}); err != nil {
			log.Printf("Error walking routes: %v", err)
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully.")
}

func getServerPort() string {
	return getEnv("PORT", "8080")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
