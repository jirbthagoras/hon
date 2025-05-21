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
		slog.Error("Error while inerting data", "err", err)
		return err
	}

	return nil
}

func (s *ProducerService) UserLogin(req RequestAuthUser) error {
	// init some vars
	var user User

	// Create a query
	query := "SELECT * FROM users WHERE email = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Query and checks if the user with such email exists or nah
	err = tx.QueryRowContext(context.Background(), query, user.Email).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("User not found", "err", err)
			return fiber.NewError(fiber.StatusBadRequest, "User with such credentials does not exist")
		}
	}

	if user.Password != req.Password {
		slog.Error("Wrong password", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Wrong password")
	}

	return nil
}
