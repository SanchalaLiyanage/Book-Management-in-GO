package handlers

import (
	"book-api/models"
	"book-api/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"             //mux = router for path parameters (like /books/{id})
	"github.com/stretchr/testify/assert" // helpful for writing readable test checks
)

type mockBookRepository struct { //A fake version of your real book repository. It avoids actual file/database calls during tests.
	books []*models.Book
}

func (m *mockBookRepository) GetAllBooks() ([]*models.Book, error) {
	return m.books, nil
}

func (m *mockBookRepository) GetBookByID(id string) (*models.Book, error) {
	for _, book := range m.books {
		if book.BookID == id {
			return book, nil
		}
	}
	return nil, repository.ErrBookNotFound
}

func (m *mockBookRepository) CreateBook(book *models.Book) error {
	for _, b := range m.books {
		if b.ISBN == book.ISBN {
			return repository.ErrDuplicateISBN
		}
	}
	m.books = append(m.books, book)
	return nil
}

func (m *mockBookRepository) UpdateBook(id string, book *models.Book) (*models.Book, error) {
	for i, b := range m.books {
		if b.BookID == id {
			for j, existing := range m.books {
				if i != j && existing.ISBN == book.ISBN {
					return nil, repository.ErrDuplicateISBN
				}
			}
			book.BookID = id
			book.CreatedAt = b.CreatedAt
			m.books[i] = book
			return book, nil
		}
	}
	return nil, repository.ErrBookNotFound
}

func (m *mockBookRepository) DeleteBook(id string) error {
	for i, book := range m.books {
		if book.BookID == id {
			m.books = append(m.books[:i], m.books[i+1:]...)
			return nil
		}
	}
	return repository.ErrBookNotFound
}

func TestBookHandler_GetBooks(t *testing.T) { //Tests the GET /books endpoint.
	repo := &mockBookRepository{
		books: []*models.Book{ //Creates a fake list of 2 books.
			{BookID: "1", Title: "Book 1"},
			{BookID: "2", Title: "Book 2"},
		},
	}
	handler := NewBookHandler(repo)

	req, err := http.NewRequest("GET", "/books", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetBooks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestBookHandler_CreateBook(t *testing.T) {
	repo := &mockBookRepository{books: []*models.Book{}}
	handler := NewBookHandler(repo)

	newBook := models.Book{
		Title:       "New Book",
		AuthorID:    "author1",
		PublisherID: "pub1",
		ISBN:        "1234567890",
		Pages:       100,
		Price:       19.99,
		Quantity:    5,
	}

	body, err := json.Marshal(newBook)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/books", bytes.NewBuffer(body))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.CreateBook(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdBook models.Book
	err = json.Unmarshal(rr.Body.Bytes(), &createdBook)
	assert.NoError(t, err)
	assert.Equal(t, "New Book", createdBook.Title)
	assert.NotEmpty(t, createdBook.BookID)
}

func TestBookHandler_GetBook(t *testing.T) {
	repo := &mockBookRepository{
		books: []*models.Book{
			{BookID: "1", Title: "Book 1"},
		},
	}
	handler := NewBookHandler(repo)

	req, err := http.NewRequest("GET", "/books/1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/books/{id}", handler.GetBook)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var book models.Book
	err = json.Unmarshal(rr.Body.Bytes(), &book)
	assert.NoError(t, err)
	assert.Equal(t, "1", book.BookID)
	assert.Equal(t, "Book 1", book.Title)
}

func TestBookHandler_UpdateBook(t *testing.T) {
	repo := &mockBookRepository{
		books: []*models.Book{
			{BookID: "1", Title: "Old Title"},
		},
	}
	handler := NewBookHandler(repo)

	updatedBook := models.Book{
		Title:       "New Title",
		AuthorID:    "author1",
		PublisherID: "pub1",
		ISBN:        "1234567890",
		Pages:       100,
		Price:       19.99,
		Quantity:    5,
	}

	body, err := json.Marshal(updatedBook)
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "/books/1", bytes.NewBuffer(body))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/books/{id}", handler.UpdateBook)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var book models.Book
	err = json.Unmarshal(rr.Body.Bytes(), &book)
	assert.NoError(t, err)
	assert.Equal(t, "New Title", book.Title)
}

func TestBookHandler_DeleteBook(t *testing.T) {
	repo := &mockBookRepository{
		books: []*models.Book{
			{BookID: "1", Title: "Book 1"},
		},
	}
	handler := NewBookHandler(repo)

	req, err := http.NewRequest("DELETE", "/books/1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/books/{id}", handler.DeleteBook)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Body.String())
}
