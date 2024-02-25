package sqlstore

import (
	"database/sql"

	"github.com/QMAwerda/http-rest-api/internal/app/store"
	_ "github.com/lib/pq" // анонимный импорт, чтобы методы в код не импортировались
)

type Store struct {
	db             *sql.DB
	UserRepository *UserRepository
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// store.User().Create() - пример обращения к репозиторию из внешнего мира
func (s *Store) User() store.UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &UserRepository{
		store: s,
	}

	return s.UserRepository
}
