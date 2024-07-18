package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	u, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", u)
}
