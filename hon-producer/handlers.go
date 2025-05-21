package main

import (
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jirbthagoras/hon/shared"
)

type ProducerHandler struct {
	Validator *validator.Validate
	Service   *ProducerService
}

func NewProducerHandler(v *validator.Validate, s *ProducerService) *ProducerHandler {
	return &ProducerHandler{Validator: v, Service: s}
}

func (h *ProducerHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/register", h.handleRegister)
	router.Post("/login", h.handleLogin)
}

func (h *ProducerHandler) handleRegister(c *fiber.Ctx) error {
	// initializing
	req := &RequestAuthUser{}
	err := c.BodyParser(req)
	if err != nil {
		slog.Error("Error while parsing body", "err", err)
		return err
	}

	// Validate
	err = h.Validator.Struct(req)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*req, err.(validator.ValidationErrors))
	}

	// Calls the Producer Service
	err = h.Service.CreateUser(*req)
	if err != nil {
		return err
	}

	// Creates a JWT token
	expiry := time.Now().Add(24 * time.Hour)
	token, err := shared.GenerateToken(req.Email, expiry)
	if err != nil {
		slog.Error("Error while generating token", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Create user success",
		"token":   token,
	})
}

func (h *ProducerHandler) handleLogin(c *fiber.Ctx) error {
	// initializing
	req := &RequestAuthUser{}
	err := c.BodyParser(req)
	if err != nil {
		slog.Error("Error while parsing body", "err", err)
		return err
	}

	// Validate
	err = h.Validator.Struct(req)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*req, err.(validator.ValidationErrors))
	}

	// Calls the Producer Service
	err = h.Service.UserLogin(*req)
	if err != nil {
		return err
	}

	expiry := time.Now().Add(24 * time.Hour)
	token, err := shared.GenerateToken(req.Email, expiry)
	if err != nil {
		slog.Error("Error while generating token", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Create user success",
		"token":   token,
	})
}
