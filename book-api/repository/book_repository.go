package repository

import (
	"book-api/models"
	"errors"
)

var (
	ErrBookNotFound  = errors.New("book not found")
	ErrDuplicateISBN = errors.New("book with this ISBN already exists")
)

type BookRepository interface {
	GetAllBooks() ([]*models.Book, error)
	GetBookByID(id string) (*models.Book, error)
	CreateBook(book *models.Book) error
	UpdateBook(id string, book *models.Book) (*models.Book, error)
	DeleteBook(id string) error
}

type FileBookRepository struct {
	store *FileStore
}

func NewBookRepository(store *FileStore) *FileBookRepository {
	return &FileBookRepository{store: store}
}

func (r *FileBookRepository) GetAllBooks() ([]*models.Book, error) {
	return r.store.ReadAll()
}

func (r *FileBookRepository) GetBookByID(id string) (*models.Book, error) {
	books, err := r.store.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, book := range books {
		if book.BookID == id {
			return book, nil
		}
	}

	return nil, ErrBookNotFound
}

func (r *FileBookRepository) CreateBook(book *models.Book) error {
	books, err := r.store.ReadAll()
	if err != nil {
		return err
	}

	for _, b := range books {
		if b.ISBN == book.ISBN {
			return ErrDuplicateISBN
		}
	}

	books = append(books, book)
	return r.store.WriteAll(books)
}

func (r *FileBookRepository) UpdateBook(id string, updatedBook *models.Book) (*models.Book, error) {
	books, err := r.store.ReadAll()
	if err != nil {
		return nil, err
	}

	for i, book := range books {
		if book.BookID == id {
			for j, b := range books {
				if i != j && b.ISBN == updatedBook.ISBN {
					return nil, ErrDuplicateISBN
				}
			}

			updatedBook.BookID = id
			updatedBook.CreatedAt = book.CreatedAt
			updatedBook.UpdatedAt = models.NewBook().UpdatedAt
			books[i] = updatedBook

			if err := r.store.WriteAll(books); err != nil {
				return nil, err
			}
			return updatedBook, nil
		}
	}

	return nil, ErrBookNotFound
}

func (r *FileBookRepository) DeleteBook(id string) error {
	books, err := r.store.ReadAll()
	if err != nil {
		return err
	}

	for i, book := range books {
		if book.BookID == id {
			books = append(books[:i], books[i+1:]...)
			return r.store.WriteAll(books)
		}
	}

	return ErrBookNotFound
}
