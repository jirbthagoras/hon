package main

import "time"

type RequestAuthUser struct {
	Id       int
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=30"`
}

type RequestCreateBook struct {
	Title      string `json:"title" validate:"required,min=6,max=50"`
	Author     string `json:"author" validate:"required"`
	TotalPages int    `json:"total_pages" validate:"required"`
}

type ResponseGetBook struct {
	Id         int                    `json:"id"`
	Title      string                 `json:"title"`
	Author     string                 `json:"author"`
	TotalPages int                    `json:"total_pages"`
	Status     string                 `json:"status"`
	Progresses []*ResponseGetProgress `json:"progresses"`
}

type RequestCreateProgress struct {
	UserId      int
	BookId      int
	UntilPage   int    `json:"until_page" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type ResponseGetProgress struct {
	Id          int       `json:"id"`
	FromPage    int       `json:"from_page"`
	UntilPage   int       `json:"until_page"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"Description"`
}
