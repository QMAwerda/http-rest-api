package store

import (
	"fmt"
	"strings"
	"testing"
)

// Этот хелпер будет возвращать нам тестовый Store, который уже будет сконфигурирован определенным образом, а также будет
// возвращать функцию, при вызове которой мы сможем очищать все таблицы, чтобы следующие тесты могли работать с уже пустой бд.

func TestStore(t *testing.T, databaseURL string) (*Store, func(...string)) {
	t.Helper()

	config := NewConfig()
	config.DatabaseURL = databaseURL
	s := New(config)
	if err := s.Open(); err != nil { // If not connected
		t.Fatal(err) // fail test
	}

	return s, func(tables ...string) {
		if len(tables) > 0 { // if we have table names in the parametrs
			if _, err := s.db.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ","))); err != nil {
				t.Fatal(err)
			}
		}

		s.Close()
	}
}
