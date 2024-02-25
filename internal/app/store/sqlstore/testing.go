package sqlstore

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// Этот хелпер будет возвращать нам тестовую БД, которая уже будет сконфигурирован определенным образом, а также будет
// возвращать функцию, при вызове которой мы сможем очищать все таблицы, чтобы следующие тесты могли работать с уже пустой бд.

func TestDB(t *testing.T, databaseURL string) (*sql.DB, func(...string)) {
	t.Helper()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", ")))
		}

		db.Close()
	}
}
