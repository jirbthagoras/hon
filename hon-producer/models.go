package main

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Book struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id"`
	Title      int    `json:"title"`
	Author     string `json:"author"`
	TotalPages string `json:"total_pages"`
	Status     string `json:"status"`
}
