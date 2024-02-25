package sqlstore

import (
	"database/sql"

	"github.com/QMAwerda/http-rest-api/internal/app/model"
	"github.com/QMAwerda/http-rest-api/internal/app/store"
)

// Repositories are responsible for work with database, they put data from the db into the models (for expamle)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil { // проверяем данные на валидность, если валидны, запускаем колбек BeforeCreate()
		return err
	}

	// Коллбэк — это функция, которая должна быть выполнена после того, как другая функция завершит работу.
	if err := u.BeforeCreate(); err != nil { // хешируем пароль
		return err
	}

	return r.store.db.QueryRow( // добавляем пользователя в БД
		"INSERT INTO users (email, encrypted_password) VALUES ($1, $2) RETURNING id",
		u.Email,             // подставляемый параметр
		u.EncryptedPassword, // подставляемый параметр
	).Scan(&u.ID)
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, email, encrypted_password FROM users WHERE email=$1",
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
	); err == sql.ErrNoRows { // отлов ошибки из postgres
		return nil, store.ErrRecordNotFound // вернем нашу ошибку
	}

	return u, nil // Тут мы вернем заполненного юзера (заполняем методом Scan())
}
