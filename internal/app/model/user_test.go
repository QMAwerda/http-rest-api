package model_test

import (
	"testing"

	"github.com/QMAwerda/http-rest-api/internal/app/model"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeCreate(t *testing.T) { // тестируем хеш функцию
	u := model.TestUser(t)
	assert.NoError(t, u.BeforeCreate()) // Кажется, она проверяет только на отстутсвие ошибок
	assert.NotEmpty(t, u.EncryptedPassword)
}

func TestUser_Validate(t *testing.T) {
	// Тут мы тестируем работу валидатора на тестовых кейсах (отсутсвие поле или некорректное поле)
	testCases := []struct {
		name    string
		u       func() *model.User
		isValid bool
	}{
		{
			name: "valid",
			u: func() *model.User {
				return model.TestUser(t)
			},
			isValid: true,
		},
		{
			name: "with encrypted password",
			u: func() *model.User {
				u := model.TestUser(t)
				u.Password = ""
				u.EncryptedPassword = "encryptedpassword" // поле не пустое

				return u
			},
			isValid: true,
		},
		{
			name: "empty email",
			u: func() *model.User {
				u := model.TestUser(t)
				u.Email = ""

				return u
			},
			isValid: false, // по нашей идее, поле email не может быть пустым
		},
		{
			name: "invalid email",
			u: func() *model.User {
				u := model.TestUser(t)
				u.Email = "invalid"

				return u
			},
			isValid: false, // по нашей идее, поле email должен содержать именно email, а не что-то другое
		},
		{
			name: "empty password",
			u: func() *model.User {
				u := model.TestUser(t)
				u.Password = ""

				return u
			},
			isValid: false,
		},
		{
			name: "short password",
			u: func() *model.User {
				u := model.TestUser(t)
				u.Password = "short"

				return u
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.u().Validate()) // ловим корректное поведение валидатора
			} else {
				assert.Error(t, tc.u().Validate()) // ловим ошибку брошенную валидатором
			}
		})
	}
}
