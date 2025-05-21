package main

import (
	"fmt"

	"github.com/jirbthagoras/hon/shared"
)

func main() {
	fmt.Println(shared.NewConfig().GetString("JWT_SECRET_KEY"))
	fmt.Println(shared.NewConfig().GetString("DB_NAME"))
	fmt.Println(shared.NewConfig().GetString("FOO"))
}
