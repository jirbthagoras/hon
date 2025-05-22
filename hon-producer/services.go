package main

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jirbthagoras/hon/shared"
)

// Creating a Service in form of struct to make it easy
type ProducerService struct {
	DB *sql.DB
}

func NewProducerService(db *sql.DB) *ProducerService {
	return &ProducerService{db}
}

func (s *ProducerService) GetUserByEmail(email string) (*User, error) {
	// init some vars
	var user User

	// Create a query
	query := "SELECT * FROM users WHERE email = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query and checks if the user with such email exists or nah
	err = tx.QueryRowContext(context.Background(), query, email).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("User not found", "err", err)
			return nil, fiber.NewError(fiber.StatusBadRequest, "User with such credentials does not exist")
		}
	}

	return &user, nil
}

func (s *ProducerService) CreateUser(req RequestAuthUser) (int, error) {
	// Create a query
	query := "INSERT INTO users (email, password) VALUES (?, ?)"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return 0, err
	}

	// Execute the query with ExecContext
	result, err := tx.ExecContext(context.Background(), query, req.Email, req.Password)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return 0, err
	}

	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("Failed to get last insert ID", "err", err)
		return 0, err
	}

	return int(id), nil
}

func (s *ProducerService) UserLogin(req RequestAuthUser) (int, error) {
	// init some vars
	user, err := s.GetUserByEmail(req.Email)
	if err != nil {
		return 0, nil
	}

	if user.Password != req.Password {
		slog.Error("Wrong password", "err", err)
		return 0, fiber.NewError(fiber.StatusBadRequest, "Wrong password")
	}

	return user.Id, nil
}

// BOOKS

func (s *ProducerService) CreateBook(userId int, req RequestCreateBook) error {
	// Crate the query
	query := "INSERT INTO books(user_id, title, author, total_pages) values(?, ?, ?, ?)"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query
	_, err = tx.ExecContext(context.Background(), query, userId, req.Title, req.Author, req.TotalPages)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	return nil
}

func (s *ProducerService) GetAllBooksByUserId(userId int) ([]*ResponseGetBook, error) {
	// Initialize var to place the book
	var books []*ResponseGetBook

	// Query

	query := "SELECT id, title, author, total_pages, status FROM books WHERE user_id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query
	rows, err := tx.QueryContext(context.Background(), query, userId)
	if err != nil {
		slog.Error("Eror while query", "err", err)
		return nil, err
	}

	// close the rows of course
	defer rows.Close()

	// Foreach-ing queried rows
	for rows.Next() {
		var book ResponseGetBook
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.TotalPages, &book.Status)
		if err != nil {
			slog.Error("Error querying", "err", err)
			return nil, err
		}
		books = append(books, &book)
	}

	return books, err
}

func (s *ProducerService) GetBookById(bookId int, userId int) (*ResponseGetBook, error) {
	// init some vars
	var book ResponseGetBook

	// Create a query
	query := "SELECT id, title, author, total_pages, status FROM books WHERE id = ? && user_id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query and checks if the user with such email exists or nah
	err = tx.QueryRowContext(context.Background(), query, bookId, userId).Scan(&book.Id, &book.Title, &book.Author, &book.TotalPages, &book.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("Book not found", "err", err)
			return nil, fiber.NewError(fiber.StatusBadRequest, "Book with such credentials does not exist")
		}
	}

	return &book, nil
}
