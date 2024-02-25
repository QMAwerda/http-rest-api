package teststore

// Для локальных тестов было решено взять хеш мапу, которая эмулирует поведение базы данных.
// Хотя вообще это странно, мне кажется, стоит локально запускать у себя бд.

import (
	"github.com/QMAwerda/http-rest-api/internal/app/model"
	"github.com/QMAwerda/http-rest-api/internal/app/store"
)

type UserRepository struct {
	store *Store
	users map[int]*model.User // ключом будет почта, а значением - указатель на пользователя
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil { // проверяем данные на валидность, если валидны, запускаем колбек BeforeCreate()
		return err
	}

	if err := u.BeforeCreate(); err != nil { // хешируем пароль
		return err
	}

	u.ID = len(r.users)
	r.users[u.ID] = u // добавляем пользователя по id

	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}
