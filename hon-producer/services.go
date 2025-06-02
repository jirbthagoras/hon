package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jirbthagoras/hon/shared"
	"github.com/rabbitmq/amqp091-go"
)

// Creating a Service in form of struct to make it easy
type ProducerService struct {
	DB   *sql.DB
	AMQP *shared.AMQP
}

func NewProducerService(db *sql.DB, amqp *shared.AMQP) *ProducerService {
	return &ProducerService{DB: db, AMQP: amqp}
}

// USERS

func (s *ProducerService) GetUser(identifier string) (*User, error) {
	// Decide whether identifier is email or id
	var field = "id"
	if strings.Contains(identifier, "@") {
		field = "email"
	}

	// init some vars
	var user User

	// Create a query
	query := fmt.Sprintf("SELECT * FROM users WHERE %s = ?", field)

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query and checks if the user with such email exists or nah
	err = tx.QueryRowContext(context.Background(), query, identifier).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("User not found", "err", err)
			return &user, fiber.NewError(fiber.StatusBadRequest, "User with such credentials does not exist")
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
	user, err := s.GetUser(req.Email)
	if err != nil {
		return 0, err
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
		return &book, err
	}

	// Query and checks if the book exist
	err = tx.QueryRowContext(context.Background(), query, bookId, userId).Scan(&book.Id, &book.Title, &book.Author, &book.TotalPages, &book.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("Book not found", "err", err)
			return &book, fiber.NewError(fiber.StatusBadRequest, "Book with such credentials does not exist")
		}
	}

	return &book, nil
}

func (s *ProducerService) DeleteBookById(bookId int, userId int) error {
	// Create a query
	query := "DELETE FROM books where id = ? && user_id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	result, err := tx.ExecContext(context.Background(), query, bookId, userId)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	// checks the affected row to make sure if there is in fact deleted book
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		slog.Error("Failed, no rows affected")
		return fiber.NewError(fiber.StatusBadRequest, "Delete failed, book probably does not exist")
	}

	return nil
}

func (s *ProducerService) SetBookStatus(bookId int, status string) error {
	if status != "reading" && status != "completed" {
		slog.Error("Status invalid")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	// Create a query
	query := "UPDATE books SET status = ? WHERE id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	result, err := tx.ExecContext(context.Background(), query, status, bookId)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	// checks the affected row to make sure if there is in fact updated book
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		slog.Error("Failed, no rows affected")
		return fiber.NewError(fiber.StatusBadRequest, "Update failed, no rows affected")
	}

	return nil
}

// PROGRESSES
func (s *ProducerService) CreateProgress(req RequestCreateProgress) error {

	// Acquire latest progress for validation purpose (make sure if the FROM_PAGE and UNTIl_PAGE is right)
	previousProgress, err := s.getLatestProgress(req.BookId)
	if err != nil {
		return err
	}

	// Acquire the book to check if the page is maxed out or not
	book, err := s.GetBookById(req.BookId, req.UserId)
	if err != nil {
		slog.Error("GetBookById")
		return err
	}

	// checks if the book already finished?
	if book.Status == "completed" {
		return fiber.NewError(fiber.StatusBadRequest, "Sorry but you're already finished your book!")
	}

	// checks if it's exceeds book page
	if req.UntilPage > book.TotalPages {
		return fiber.NewError(fiber.StatusBadRequest, "Until Page exceeds book's page")
	}

	// checks if it's maxed out or nah
	if book.TotalPages == req.UntilPage {
		// Set the book status to completed
		err = s.SetBookStatus(book.Id, "completed")
		if err != nil {
			slog.Error("Calls setbookstatus")
			return err
		}
	}

	// Checks if the progress fulfilled a goal
	goals, err := s.GetAllGoals(req.BookId, req.UserId)
	if err != nil {
		slog.Error("Calls GetAllGoals")
		return err
	}

	// Checks all goals
	for _, goal := range goals {
		// Skip if goal's status finished
		if goal.Status != "in-progress" {
			continue
		}
		if req.UntilPage >= goal.TargetPage {
			// Set the goal status
			err = s.SetGoalStatus("finished", goal.Id)
			if err != nil {
				return err
			}

			user, err := s.GetUser(strconv.Itoa(req.UserId))
			if err != nil {
				return err
			}

			msg := &shared.Msg{
				Id:         goal.Id,
				Email:      user.Email,
				Name:       goal.Name,
				BookTitle:  book.Title,
				TargetPage: goal.TargetPage,
				ExpiredAt:  goal.ExpiredAt,
			}

			body, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			// Send the message to exchange
			err = s.sendGoalMessage(body)
			if err != nil {
				return err
			}
		}
	}

	// If latest progress exist, take it's until_page as new progress' from_page.
	// if not, then it's the first page.
	fromPage := 0
	if previousProgress != nil {
		fromPage = previousProgress.UntilPage
	}

	if previousProgress.UntilPage >= req.UntilPage {
		return fiber.NewError(fiber.StatusBadRequest, "Current until_page is lesser or same as previous until_page, no improvement")
	}

	// Create a query
	query := "INSERT INTO progresses(book_id, from_page, until_page, description) VALUES (?, ?, ?, ?)"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	_, err = tx.ExecContext(context.Background(), query, req.BookId, fromPage, req.UntilPage, req.Description)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	return nil
}

func (s *ProducerService) sendGoalMessage(body []byte) error {
	// Make agent
	agent, err := shared.NewAgent(s.AMQP, context.Background())
	if err != nil {
		return err
	}

	// Make agent work! Publish a message
	err = agent.Publish(amqp091.Publishing{
		Headers: amqp091.Table{
			"sample": "value",
		},
		Body: body,
	}, "goal_exchange", "goal")

	if err != nil {
		return err
	}

	return nil
}

func (s *ProducerService) getLatestProgress(bookId int) (*Progress, error) {
	// init some vars
	var progress Progress

	// Create a query
	query := "SELECT * FROM progresses WHERE book_id = ? ORDER BY created_at DESC LIMIT 1"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query and checks if the progress exists
	err = tx.QueryRowContext(context.Background(), query, bookId).Scan(
		&progress.Id,
		&progress.BookId,
		&progress.FromPage,
		&progress.UntilPage,
		&progress.Description,
		&progress.CreatedAt,
	)
	// if its empty, that's OKAY! Cuz it's the first progress
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Info("This is the first progress of book!")
			return &progress, nil
		}
	}

	return &progress, nil
}

func (s *ProducerService) GetAllProgressByBookId(bookId int) ([]*ResponseGetProgress, error) {
	var progresses []*ResponseGetProgress

	// Create a query
	query := "SELECT id, from_page, until_page, created_at, description FROM progresses WHERE book_id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query
	rows, err := tx.QueryContext(context.Background(), query, bookId)
	if err != nil {
		slog.Error("Eror while query", "err", err)
		return nil, err
	}

	// close the rows of course
	defer rows.Close()

	// Foreach-ing queried rows
	for rows.Next() {
		var progress ResponseGetProgress
		err := rows.Scan(&progress.Id, &progress.FromPage, &progress.UntilPage, &progress.CreatedAt, &progress.Description)
		if err != nil {
			slog.Error("Error querying", "err", err)
			return nil, err
		}
		progresses = append(progresses, &progress)
	}

	return progresses, err
}

func (s *ProducerService) DeleteLatestProgress(bookId int, userId int) error {
	// fetch latest progress
	progress, err := s.getLatestProgress(bookId)
	if err != nil {
		return err
	}

	// checks if the time of deletion is still valid
	expiry := progress.CreatedAt.Add(1 * time.Minute)
	if !time.Now().Before(expiry) {
		slog.Error("Cannot delete progress that already created in 1 minute")
		return fiber.NewError(fiber.StatusBadRequest, "Cannot delete progress that already created past 1 minute")
	}

	// Get book to access the status
	book, err := s.GetBookById(bookId, userId)
	if err != nil {
		return err
	}

	// if the book status completed, change the status to reading
	if book.Status == "completed" {
		err := s.SetBookStatus(bookId, "reading")
		if err != nil {
			return err
		}
	}

	// query
	query := "DELETE FROM progresses WHERE id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	_, err = tx.ExecContext(context.Background(), query, progress.Id)
	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	return nil
}

// GOALS

func (s *ProducerService) CreateGoal(req *RequestCreateGoal) error {
	// Checks if the user hold the book
	book, err := s.GetBookById(req.BookId, req.UserId)
	if err != nil {
		return err
	}

	// Checks if the book is already finished
	if book.Status == "completed" {
		return fiber.NewError(fiber.StatusBadRequest, "Book already finished, nothing to chase bro")
	}

	// acquire latest progress
	progress, err := s.getLatestProgress(book.Id)
	if err != nil {
		return err
	}

	// Checks if the target page exceeds book latest progress.
	if progress.UntilPage >= req.TargetPage {
		return fiber.NewError(fiber.StatusBadRequest, "Your target already fulfilled or maybe exceeds your latest progress")
	}

	// Validate the expired_time
	if !time.Now().Before(req.ExpiredAt) {
		return fiber.NewError(fiber.StatusBadRequest, "Expired Time is invalid")
	}

	// Create a query
	query := "INSERT INTO goals (book_id, user_id, name, target_page, expired_at) VALUES (?, ?, ?, ?, ?)"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	result, err := tx.ExecContext(context.Background(), query,
		req.BookId,
		req.UserId,
		req.Name,
		req.TargetPage,
		req.ExpiredAt)

	if err != nil {
		slog.Error("Error while inserting data", "err", err)
		return err
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Find a user first to get the email
	user, err := s.GetUser(strconv.Itoa(req.UserId))
	if err != nil {
		return err
	}

	// Crafts a body
	msg := &shared.Msg{
		Id:         int(lastInsertId),
		Email:      user.Email,
		Name:       req.Name,
		BookTitle:  book.Title,
		TargetPage: req.TargetPage,
		ExpiredAt:  req.ExpiredAt,
	}

	// Craft a delay time first, expired_time - now
	delay := time.Until(msg.ExpiredAt)
	slog.Info(strconv.Itoa(int(delay.Milliseconds())))

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Make agent to send delayed message to queue.
	agent, err := shared.NewAgent(s.AMQP, context.Background())
	if err != nil {
		return err
	}

	// Publish the message
	err = agent.Publish(amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		ContentType:  "application/json",
		Body:         body,
		Headers: amqp091.Table{
			"x-delay": delay.Milliseconds(), // delay in milliseconds
		},
	}, "goal_exchange", "deadline")

	if err != nil {
		return err
	}

	return nil
}

func (s *ProducerService) GetAllGoals(bookId int, userId int) ([]*ResponseGetGoal, error) {
	// Initialize var to place the book
	var goals []*ResponseGetGoal

	// Query
	query := "SELECT id, name, target_page, status, expired_at FROM goals WHERE book_id = ? && user_id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return nil, err
	}

	// Query
	rows, err := tx.QueryContext(context.Background(), query, bookId, userId)
	if err != nil {
		slog.Error("Eror while query", "err", err)
		return nil, err
	}

	// close the rows of course
	defer rows.Close()

	for rows.Next() {
		var goal ResponseGetGoal
		err := rows.Scan(
			&goal.Id,
			&goal.Name,
			&goal.TargetPage,
			&goal.Status,
			&goal.ExpiredAt,
		)
		if err != nil {
			slog.Error("Error Querying")
			return goals, err
		}
		goals = append(goals, &goal)
	}

	return goals, nil
}

func (s *ProducerService) SetGoalStatus(status string, goalId int) error {
	if status != "finished" && status != "expired" {
		slog.Error("Status invalid")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	// Create a query
	query := "UPDATE goals SET status = ? WHERE id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return err
	}

	// Execute the query with ExecContext
	result, err := tx.ExecContext(context.Background(), query, status, goalId)
	if err != nil {
		slog.Error("Error while inserting data setgoal", "err", err)
		return err
	}

	// checks the affected row to make sure if there is in fact updated book
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		slog.Error("Failed, no rows affected")
		return fiber.NewError(fiber.StatusBadRequest, "Update failed, no rows affected")
	}

	return nil
}
