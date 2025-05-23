package main

import "time"

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Book struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	TotalPages string `json:"total_pages"`
	Status     string `json:"status"`
}

type Progress struct {
	Id          int       `json:"id"`
	BookId      int       `json:"book_id"`
	FromPage    int       `json:"from_page" validate:"required"`
	UntilPage   int       `json:"until_page" validate:"required"`
	Description string    `json:"description" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}
