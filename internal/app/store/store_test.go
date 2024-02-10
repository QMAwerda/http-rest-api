package store_test

import (
	"os"
	"testing"
)

var (
	databaseURL string
)

// Эта фукнкция вызовется один раз перед всеми тестами
func TestMain(m *testing.M) {
	// переменная окружения - объект, содержащий текстовую инфу, которую могут использовать запускаемые программы.
	databaseURL = os.Getenv("DATABASE_URL") // вроде возвращает значение из переменной окружения
	if databaseURL == "" {
		databaseURL = "host=localhost dbname=restapi_test sslmode=disable user=postgres password=postgres"
	}

	os.Exit(m.Run())
}
