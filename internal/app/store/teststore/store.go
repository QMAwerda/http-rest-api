package teststore

import (
	"github.com/QMAwerda/http-rest-api/internal/app/model"
	"github.com/QMAwerda/http-rest-api/internal/app/store"
)

type Store struct {
	UserRepository *UserRepository
}

func New() *Store {
	return &Store{}
}

// store.User().Create() - пример обращения к репозиторию из внешнего мира
func (s *Store) User() store.UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &UserRepository{
		store: s,
		users: make(map[int]*model.User),
	}

	return s.UserRepository
}
