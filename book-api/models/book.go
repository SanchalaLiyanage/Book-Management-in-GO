package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	BookID          string    `json:"bookId"`
	AuthorID        string    `json:"authorId"`
	PublisherID     string    `json:"publisherId"`
	Title           string    `json:"title"`
	PublicationDate string    `json:"publicationDate"`
	ISBN            string    `json:"isbn"`
	Pages           int       `json:"pages"`
	Genre           string    `json:"genre"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	Quantity        int       `json:"quantity"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func NewBook() *Book {
	return &Book{
		BookID:    uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (b *Book) Validate() error {
	var errs []string

	if b.Title == "" {
		errs = append(errs, "title is required")
	}
	if b.AuthorID == "" {
		errs = append(errs, "authorId is required")
	}
	if b.PublisherID == "" {
		errs = append(errs, "publisherId is required")
	}
	if b.ISBN == "" {
		errs = append(errs, "isbn is required")
	}
	if b.Pages <= 0 {
		errs = append(errs, "pages must be positive")
	}
	if b.Price <= 0 {
		errs = append(errs, "price must be positive")
	}
	if b.Quantity < 0 {
		errs = append(errs, "quantity cannot be negative")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}
