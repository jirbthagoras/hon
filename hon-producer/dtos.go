package main

type RequestAuthUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=30"`
}

type RequestCreateBook struct {
	Title      string `json:"title" validate:"required,min=6,max=50"`
	Author     string `json:"author" validate:"required"`
	TotalPages int    `json:"total_pages" validate:"required"`
	Email      string
}
