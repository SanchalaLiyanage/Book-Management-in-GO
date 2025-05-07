package handlers

import (
	"book-api/models"
	"book-api/repository"
	"encoding/json"
	"net/http" //core HTTP utilities.
	"strconv"  //string to number conversion.

	"github.com/gorilla/mux"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo} //Constructor function to initialize and return a BookHandler instance.
}

func (h *BookHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	limit, offset := getPaginationParams(r)

	books, err := h.repo.GetAllBooks()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	start := offset
	if start > len(books) {
		start = len(books)
	}
	end := start + limit
	if end > len(books) {
		end = len(books)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":   books[start:end],
		"total":  len(books),
		"limit":  limit,
		"offset": offset,
	})
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := book.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	newBook := models.NewBook()
	newBook.AuthorID = book.AuthorID
	newBook.PublisherID = book.PublisherID
	newBook.Title = book.Title
	newBook.PublicationDate = book.PublicationDate
	newBook.ISBN = book.ISBN
	newBook.Pages = book.Pages
	newBook.Genre = book.Genre
	newBook.Description = book.Description
	newBook.Price = book.Price
	newBook.Quantity = book.Quantity

	if err := h.repo.CreateBook(newBook); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, newBook)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) { //to get single book by id
	vars := mux.Vars(r)
	id := vars["id"]

	book, err := h.repo.GetBookByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Book not found")
		return
	}

	respondWithJSON(w, http.StatusOK, book)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updatedBook models.Book
	if err := json.NewDecoder(r.Body).Decode(&updatedBook); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := updatedBook.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.repo.UpdateBook(id, &updatedBook)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.repo.DeleteBook(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getPaginationParams(r *http.Request) (limit, offset int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit = 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset = 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	return limit, offset
}

func respondWithError(w http.ResponseWriter, code int, message string) { //Sends a JSON error response to the client.
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) { //Sends any data as a JSON response with a status code.
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
