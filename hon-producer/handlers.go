package main

import (
	"errors"
	"log/slog"
	"strconv"
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
	auth := router.Group("/auth")
	auth.Post("/register", h.handleRegister)
	auth.Post("/login", h.handleLogin)

	book := router.Group("/book")
	book.Use(shared.TokenMiddleware)
	book.Post("/", h.handleAddBook)
	book.Get("/", h.handleGetBook)
	book.Get("/:id", h.handleGetBookById)
	book.Delete("/:id", h.handleDeleteBookById)

	progress := router.Group("/progress")
	progress.Use(shared.TokenMiddleware)
	progress.Post("/:id", h.handleCreateProgress)
	progress.Delete("/:id", h.handleCancelProgress)

	goals := router.Group("/goal")
	goals.Use(shared.TokenMiddleware)
	goals.Post("/", h.handleCreateGoal)
	goals.Get("/", h.handleGetAllGoal)
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
	id, err := h.Service.CreateUser(*req)
	if err != nil {
		return err
	}

	// Creates a JWT token
	expiry := time.Now().Add(24 * time.Hour)
	token, err := shared.GenerateToken(id, expiry)
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
	userId, err := h.Service.UserLogin(*req)
	if err != nil {
		return err
	}

	expiry := time.Now().Add(24 * time.Hour)
	token, err := shared.GenerateToken(userId, expiry)
	if err != nil {
		slog.Error("Error while generating token", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login Success, here's your token",
		"token":   token,
	})
}

func (h *ProducerHandler) handleAddBook(c *fiber.Ctx) error {
	// initializing
	req := &RequestCreateBook{}

	// Getting subject (which is user_id) from token to inject it into service.
	id, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}

	// parse the body
	err = c.BodyParser(req)
	if err != nil {
		slog.Error("Error while parsing body", "err", err)
		return err
	}

	// validate
	err = h.Validator.Struct(req)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*req, err.(validator.ValidationErrors))
	}

	// calling the service, look inside service for more detailed code
	err = h.Service.CreateBook(id, *req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Create book success",
	})
}

func (h *ProducerHandler) handleGetBook(c *fiber.Ctx) error {
	// Getting subject (which is user_id) from token to inject it into service.
	id, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}

	// Calling the service
	books, err := h.Service.GetAllBooksByUserId(id)
	if err != nil {
		slog.Error("Error while calling service", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Query Books Success",
		"books":   books,
	})
}

func (h *ProducerHandler) handleGetBookById(c *fiber.Ctx) error {
	// Taking id from params
	bookId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	// Getting subject (which is user_id) from token to inject it into service.
	userId, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}

	// calls service
	book, err := h.Service.GetBookById(bookId, userId)
	if err != nil {
		slog.Error("Error while executing service: GetBook", "err", err)
		return err
	}

	// calls service for progresses
	progresses, err := h.Service.GetAllProgressByBookId(bookId)
	if err != nil {
		slog.Error("Error while executing service: GetProgress", "err", err)
		return err
	}

	book.Progresses = progresses

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Query Book Success",
		"books":   book,
	})
}

func (h *ProducerHandler) handleDeleteBookById(c *fiber.Ctx) error {
	// Taking id from params
	bookId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	// Getting subject (which is user_id) from token to inject it into service.
	userId, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}

	// calls service
	err = h.Service.DeleteBookById(bookId, userId)
	if err != nil {
		slog.Error("Error while executing service", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Book deleted",
	})
}

func (h *ProducerHandler) handleCreateProgress(c *fiber.Ctx) error {
	// Init some vars
	req := &RequestCreateProgress{}

	// Taking id from params
	bookId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	req.BookId = bookId

	// Getting subject (which is user_id) from token to inject it into service.
	userId, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}
	req.UserId = userId

	// parse the body
	err = c.BodyParser(req)
	if err != nil {
		slog.Error("Error while parsing body", "err", err)
		return err
	}

	// Validate
	err = h.Validator.Struct(req)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*req, err.(validator.ValidationErrors))
	}

	// calls service
	err = h.Service.CreateProgress(*req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Progress successfully created",
	})

}

func (h *ProducerHandler) handleCancelProgress(c *fiber.Ctx) error {
	// Taking id from params
	bookId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	// Getting subject (which is user_id) from token to inject it into service.
	userId, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}

	// calls service
	err = h.Service.DeleteLatestProgress(bookId, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Latest progress successfully deleted",
	})
}

func (h *ProducerHandler) handleCreateGoal(c *fiber.Ctx) error {
	// Init some var
	req := &RequestCreateGoal{}

	// Parse the payload
	err := c.BodyParser(req)
	if err != nil {
		slog.Error("Error while parsing body", "err", err)
		return err
	}

	// Getting subject (which is user_id) from token to inject it into service.
	userId, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}
	req.UserId = userId

	// calls service
	err = h.Service.CreateGoal(req)
	if err != nil {
		slog.Error("Something wrong when called the service")
		return err
	}

	// Returns
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Goal created successfully",
	})
}

func (h *ProducerHandler) handleGetAllGoal(c *fiber.Ctx) error {
	// Getting subject (which is user_id) from token to inject it into service.
	id, err := shared.GetSubjectFromToken(c)
	if err != nil {
		slog.Error("Error while getting token", "err", err)
		return err
	}

	// Calling the service
	books, err := h.Service.GetAllGoals(id)
	if err != nil {
		slog.Error("Error while calling service", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Query Goals Success",
		"books":   books,
	})
}
