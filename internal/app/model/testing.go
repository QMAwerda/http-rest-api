package model

import "testing"

// Тут нет тестов, это хелперы для тестов, они сократят наш код

// Пишем хелпер, который будет возвращать пользователя с уже валидным набором данных
// Тогда в тестах не придется каждый раз вводить ему поля
func TestUser(t *testing.T) *User {
	return &User{
		Email:    "user@example.org",
		Password: "password",
	}
}
