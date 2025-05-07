package handlers

import (
	"book-api/models"
	"book-api/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSearchRepositoryForTesting struct {
	books []*models.Book
}

func (m *mockSearchRepositoryForTesting) GetAllBooks() ([]*models.Book, error) {
	return m.books, nil
}

func (m *mockSearchRepositoryForTesting) GetBookByID(id string) (*models.Book, error) {
	for _, book := range m.books {
		if book.BookID == id {
			return book, nil
		}
	}
	return nil, repository.ErrBookNotFound
}

func (m *mockSearchRepositoryForTesting) CreateBook(book *models.Book) error {
	m.books = append(m.books, book)
	return nil
}

func (m *mockSearchRepositoryForTesting) UpdateBook(id string, book *models.Book) (*models.Book, error) {
	for i, b := range m.books {
		if b.BookID == id {
			book.BookID = id
			m.books[i] = book
			return book, nil
		}
	}
	return nil, repository.ErrBookNotFound
}

func (m *mockSearchRepositoryForTesting) DeleteBook(id string) error {
	for i, book := range m.books {
		if book.BookID == id {
			m.books = append(m.books[:i], m.books[i+1:]...)
			return nil
		}
	}
	return repository.ErrBookNotFound
}

func TestExecuteBookSearchEndpoint(t *testing.T) {
	testCases := []struct {
		name               string
		query              string
		mockBooks          []*models.Book
		expectedMatchCount int
		expectedStatusCode int
		expectedErrorMsg   string
	}{
		{
			name:  "successful_title_match",
			query: "gatsby",
			mockBooks: []*models.Book{
				{
					BookID:      "1",
					Title:       "The Great Gatsby",
					Description: "A classic novel",
					AuthorID:    "F. Scott Fitzgerald",
					Genre:       "Classic",
				},
				{
					BookID:      "2",
					Title:       "To Kill a Mockingbird",
					Description: "Another classic",
					AuthorID:    "Harper Lee",
					Genre:       "Fiction",
				},
			},
			expectedMatchCount: 1,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:  "successful_description_match",
			query: "classic",
			mockBooks: []*models.Book{
				{
					BookID:      "1",
					Title:       "Book 1",
					Description: "A classic novel",
				},
				{
					BookID:      "2",
					Title:       "Book 2",
					Description: "Modern literature",
				},
			},
			expectedMatchCount: 1,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "no_matches_found",
			query:              "nonexistent",
			mockBooks:          []*models.Book{},
			expectedMatchCount: 0,
			expectedStatusCode: http.StatusNotFound,
			expectedErrorMsg:   "No books found matching your search criteria",
		},
		{
			name:               "empty_query_rejected",
			query:              "",
			mockBooks:          []*models.Book{},
			expectedMatchCount: 0,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorMsg:   "Search query cannot be empty",
		},
		{
			name:               "short_query_rejected",
			query:              "a",
			mockBooks:          []*models.Book{},
			expectedMatchCount: 0,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorMsg:   "Search query must be at least 2 characters long",
		},
		{
			name:  "case_insensitive_matching",
			query: "GREAT",
			mockBooks: []*models.Book{
				{
					BookID:      "1",
					Title:       "The Great Gatsby",
					Description: "A classic novel",
				},
			},
			expectedMatchCount: 1,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:  "match_in_author_field",
			query: "Fitzgerald",
			mockBooks: []*models.Book{
				{
					BookID:      "1",
					Title:       "The Great Gatsby",
					Description: "A classic novel",
					AuthorID:    "F. Scott Fitzgerald",
				},
			},
			expectedMatchCount: 1,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:  "match_in_genre_field",
			query: "Classic",
			mockBooks: []*models.Book{
				{
					BookID:      "1",
					Title:       "The Great Gatsby",
					Description: "A novel",
					Genre:       "Classic Literature",
				},
			},
			expectedMatchCount: 1,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockSearchRepositoryForTesting{books: tc.mockBooks}
			handler := NewSearchHandler(repo)

			req, err := http.NewRequest("GET", "/books/search?q="+tc.query, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ExecuteBookSearch(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			if tc.expectedStatusCode == http.StatusOK {
				var books []models.Book
				err := json.NewDecoder(rr.Body).Decode(&books)
				assert.NoError(t, err)
				assert.Len(t, books, tc.expectedMatchCount)
			} else if tc.expectedErrorMsg != "" {
				var errorResponse struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(rr.Body).Decode(&errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse.Error, tc.expectedErrorMsg)
			}
		})
	}
}

func TestConcurrentSearchImplementation(t *testing.T) {
	testScenarios := []struct {
		name               string
		bookCount          int
		searchTerm         string
		expectedMatchCount int
	}{
		{
			name:               "small_dataset_sequential",
			bookCount:          15,
			searchTerm:         "special",
			expectedMatchCount: 1,
		},
		{
			name:               "medium_dataset_concurrent",
			bookCount:          50,
			searchTerm:         "special",
			expectedMatchCount: 2,
		},
		{
			name:               "large_dataset_concurrent",
			bookCount:          200,
			searchTerm:         "special",
			expectedMatchCount: 3,
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			books := generateTestBooks(scenario.bookCount)
			// Add special books that should match
			books = append(books, &models.Book{
				BookID:      "special-1",
				Title:       "Special Book 1",
				Description: "Contains special keyword",
			})
			if scenario.bookCount > 50 {
				books = append(books, &models.Book{
					BookID:      "special-2",
					Title:       "Special Book 2",
					Description: "Another special book",
				})
			}
			if scenario.bookCount > 150 {
				books = append(books, &models.Book{
					BookID:      "special-3",
					Title:       "Special Book 3",
					Description: "Yet another special book",
				})
			}

			repo := &mockSearchRepositoryForTesting{books: books}
			handler := NewSearchHandler(repo)

			startTime := time.Now()
			results := handler.searchBooks(books, scenario.searchTerm)
			duration := time.Since(startTime)

			assert.Len(t, results, scenario.expectedMatchCount)
			t.Logf("%s: Searched %d books in %v, found %d matches",
				scenario.name, len(books), duration, len(results))
		})
	}
}

func TestSearchTermMatchingLogic(t *testing.T) {
	testCases := []struct {
		name          string
		book          *models.Book
		searchTerm    string
		shouldMatch   bool
		matchingField string
	}{
		{
			name: "title_match",
			book: &models.Book{
				Title:       "The Great Gatsby",
				Description: "A novel",
			},
			searchTerm:    "gatsby",
			shouldMatch:   true,
			matchingField: "title",
		},
		{
			name: "description_match",
			book: &models.Book{
				Title:       "Book 1",
				Description: "A classic novel",
			},
			searchTerm:    "classic",
			shouldMatch:   true,
			matchingField: "description",
		},
		{
			name: "author_match",
			book: &models.Book{
				Title:    "Book 1",
				AuthorID: "F. Scott Fitzgerald",
			},
			searchTerm:    "fitzgerald",
			shouldMatch:   true,
			matchingField: "author",
		},
		{
			name: "genre_match",
			book: &models.Book{
				Title: "Book 1",
				Genre: "Science Fiction",
			},
			searchTerm:    "fiction",
			shouldMatch:   true,
			matchingField: "genre",
		},
		{
			name: "no_match",
			book: &models.Book{
				Title:       "Book 1",
				Description: "A novel",
			},
			searchTerm:    "poetry",
			shouldMatch:   false,
			matchingField: "",
		},
		{
			name: "partial_word_match",
			book: &models.Book{
				Title: "The Great Gatsby",
			},
			searchTerm:    "gat",
			shouldMatch:   true,
			matchingField: "title",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := containsSearchTerm(tc.book, strings.ToLower(tc.searchTerm))
			assert.Equal(t, tc.shouldMatch, matches)
		})
	}
}

func TestSearchQueryValidation(t *testing.T) {
	validationTests := []struct {
		name        string
		query       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid_query",
			query:       "valid",
			shouldError: false,
		},
		{
			name:        "empty_query",
			query:       "",
			shouldError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "short_query",
			query:       "a",
			shouldError: true,
			errorMsg:    "at least 2 characters",
		},
		{
			name:        "long_query",
			query:       strings.Repeat("a", 101),
			shouldError: true,
			errorMsg:    "max 100 characters",
		},
	}

	for _, vt := range validationTests {
		t.Run(vt.name, func(t *testing.T) {
			err := validateSearchQuery(vt.query)
			if vt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), vt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func generateTestBooks(count int) []*models.Book {
	var books []*models.Book
	for i := 0; i < count; i++ {
		books = append(books, &models.Book{
			BookID:      string(rune(i)),
			Title:       "Book " + string(rune(i)),
			Description: "Description " + string(rune(i)),
		})
	}
	return books
}
