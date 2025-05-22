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

func (s *ProducerService) CreateUser(req RequestAuthUser) error {
	// Create a query
	query := "INSERT INTO users (email, password) VALUES (?, ?)"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	_, err = tx.ExecContext(context.Background(), query, req.Email, req.Password)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	return nil
}

func (s *ProducerService) UserLogin(req RequestAuthUser) error {
	// init some vars
	user, err := s.getUserByEmail(req.Email)
	if err != nil {
		return err
	}

	if user.Password != req.Password {
		slog.Error("Wrong password", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Wrong password")
	}

	return nil
}

func (s *ProducerService) CreateBook(req RequestCreateBook) error {
	// Get user id
	user, err := s.getUserByEmail(req.Email)
	if err != nil {
		return err
	}

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
	_, err = tx.ExecContext(context.Background(), query, user.Id, req.Title, req.Author, req.TotalPages)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	return nil
}

func (s *ProducerService) getUserByEmail(email string) (*User, error) {
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
