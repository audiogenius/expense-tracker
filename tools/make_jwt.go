package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: make_jwt <telegram_id>")
		os.Exit(2)
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Println("JWT_SECRET not set")
		os.Exit(2)
	}
	id := os.Args[1]
	// try to parse numeric id to emit number sub claim
	var sub any = id
	if v, err := strconv.ParseInt(id, 10, 64); err == nil {
		sub = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": sub, "exp": time.Now().Add(24 * time.Hour).Unix(), "iat": time.Now().Unix()})
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("sign err", err)
		os.Exit(2)
	}
	fmt.Println(s)
}
