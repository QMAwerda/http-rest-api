package model

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"golang.org/x/crypto/bcrypt"
)

// the model has the same database fields
type User struct {
	ID                int
	Email             string
	Password          string // сюда будет вводиться пароль от ползователя в открытом виде
	EncryptedPassword string
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Email, validation.Required, is.Email),
		// validation.Field(что валидируем, правило валидации)
		// в правиле валидации у нас Required - т.е. обязательное поле. is.Email - проверим, что поле в правильном формате email
		// вторая проверка идет из пакета is, той же v4 нашего валидатора
		validation.Field(&u.Password, validation.By(requiredIf(u.EncryptedPassword == "")), validation.Length(6, 100)), // минимальная длина 6, максимальная 100
	) // через validation.By() можно добавить кастомные валидации. Т.е. при получении пароля из бд мы получим хеш. Тогда
	// структура пользователя получится неполной (поле пароль - пустое), поэтому нельзя оставлять Required. Мы пишем кастомный валидатор
}

// Будем хешировать пароль перед добавлением пользователя в БД
func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 { // мы в валидаторе сверху проверяем длину пароля, так что это лишняя проверка
		enc, err := encryptedString(u.Password)
		if err != nil {
			return err
		}

		u.EncryptedPassword = enc
	}

	return nil
}

func encryptedString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
