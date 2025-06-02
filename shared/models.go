package shared

import "time"

type Msg struct {
	Id         int       `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	BookTitle  string    `json:"book_title"`
	TargetPage int       `json:"target_page"`
	ExpiredAt  time.Time `json:"expired_at"`
}
