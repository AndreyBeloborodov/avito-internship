package integration_tests

import (
	"fmt"
	"os"
)

// GetTestDSN возвращает DSN для тестовой базы
func GetTestDSN() string {
	//host := "localhost"
	//user := "postgres"
	//password := "0000"
	//dbname := "shop_test"
	//port := "5433"

	host := os.Getenv("TEST_DATABASE_HOST")
	user := os.Getenv("TEST_DATABASE_USER")
	password := os.Getenv("TEST_DATABASE_PASSWORD")
	dbname := os.Getenv("TEST_DATABASE_NAME")
	port := os.Getenv("TEST_DATABASE_PORT")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
}

func GetTestServerPort() string {
	return ":8081" // os.Getenv("TEST_SERVER_PORT")
}
