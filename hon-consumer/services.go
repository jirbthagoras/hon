package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"

	"github.com/jirbthagoras/hon/shared"
	"github.com/rabbitmq/amqp091-go"
)

type ConsumerService struct {
	Mailer *Mailer
	DB     *sql.DB
}

// return new consumer
func NewConsumerService(mailer *Mailer, DB *sql.DB) *ConsumerService {
	return &ConsumerService{
		Mailer: mailer,
		DB:     DB,
	}
}

//go:embed templates/congratulation.html
var Congratulation string

// service to Send Congrats msg
func (s *ConsumerService) SendGoalEmail(msg *amqp091.Delivery) error {
	// Create var to contain the message
	goalMsg := &shared.Msg{}

	// Parse the json cihuy
	err := json.Unmarshal(msg.Body, goalMsg)
	if err != nil {
		return err
	}

	// parse the html template
	templ, err := template.New("congratulation").Parse(Congratulation)
	if err != nil {
		return err
	}

	// Inject the msg to the templ var
	var body bytes.Buffer
	if err := templ.Execute(&body, goalMsg); err != nil {
		return err
	}

	// Make the email data that will be injected to Mailer
	emailData := SendMail{
		To:      goalMsg.Email,
		Subject: "Hon Goal Completed",
		Body:    body.String(),
	}

	if err := s.Mailer.SendMail(&emailData); err != nil {
		return err
	}

	slog.Info("Email sent successfully", "to", goalMsg.Email, "subject", emailData.Subject)

	return nil
}

//go:embed templates/deadline.html
var Deadline string

// servuce to send Failed msg
func (s *ConsumerService) SendDeadlineEmail(msg *amqp091.Delivery) error {
	// Create var to contain the message
	deadlineMsg := &shared.Msg{}

	// Parse the json cihuy
	err := json.Unmarshal(msg.Body, deadlineMsg)
	if err != nil {
		return err
	}

	// Checks the goal first, is it finished or in-progress?
	status, err := s.checkGoal(deadlineMsg)
	if err != nil {
		return err
	}

	// Checks whether if not finished, then it will be updated to expired
	if status == "finished" {
		return errors.New("The Goal is finished, nothing to do")
	}

	err = s.SetGoalStatus("expired", deadlineMsg.Id)
	if err != nil {
		return err
	}

	// parse the html template
	templ, err := template.New("deadline").Parse(Deadline)
	if err != nil {
		return err
	}

	// Inject the msg to the templ var
	var body bytes.Buffer
	if err := templ.Execute(&body, deadlineMsg); err != nil {
		return err
	}

	// Make the email data that will be injected to Mailer
	emailData := SendMail{
		To:      deadlineMsg.Email,
		Subject: "Hon Goal Expired",
		Body:    body.String(),
	}

	if err := s.Mailer.SendMail(&emailData); err != nil {
		return err
	}

	slog.Info("Email sent successfully", "to", deadlineMsg.Email, "subject", emailData.Subject)

	return nil
}

func (s *ConsumerService) checkGoal(msg *shared.Msg) (string, error) {
	// Init var
	var status string

	// make a query
	query := "SELECT status FROM goals WHERE id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		slog.Error("Commit Rollback error", "err", err)
		return status, err
	}

	// Query and checks if the book exist
	err = tx.QueryRowContext(context.Background(), query, msg.Id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("Book not found", "err", err)
			return status, err
		}

	}
	if err != nil {
		return status, err
	}

	return status, nil
}

func (s *ConsumerService) SetGoalStatus(status string, goalId int) error {
	if status != "finished" && status != "expired" {
		slog.Error("Status invalid")
		return errors.New("unknown Status injected to function")
	}

	// Create a query
	query := "UPDATE goals SET status = ? WHERE id = ?"

	// tx stuffs
	tx, err := s.DB.Begin()
	defer shared.CommitOrRollback(tx, err)
	if err != nil {
		return errors.New("something wrong when starting transaction SetGoalStatus")
	}

	// Execute the query with ExecContext
	result, err := tx.ExecContext(context.Background(), query, status, goalId)
	if err != nil {
		return errors.New("something wrong when Executing SetGoalStatus")
	}

	// checks the affected row to make sure if there is in fact updated book
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("failed, no rows affected while updating data Goal")
	}

	return nil
}
