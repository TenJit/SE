package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func EnvMongoURI() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv("MONGOURI")
}

func JWTSecret() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv("JWT_SECRET")
}

func JWTCookieExpire() int {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	expireStr := os.Getenv("JWT_COOKIE_EXPIRE")
	expireInt, err := strconv.Atoi(expireStr)
	if err != nil {
		log.Fatal("Error converting JWT_Cookie_Expire to integer")
	}
	return expireInt
}
