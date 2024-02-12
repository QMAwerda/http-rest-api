package model

import validation "github.com/go-ozzo/ozzo-validation/v4"

// Будем вызывать валидатор Required, если поле true (а там где его вызываем, будем передавать true, если хеш пароля пустой)
// Таким образом если хеш не пустой (получили из бд), тогда мы в валидаторе не проверяем поле Password на Required
func requiredIf(cond bool) validation.RuleFunc {
	return func(value interface{}) error { // тут не очень понял с интерфейсом
		if cond {
			return validation.Validate(value, validation.Required)
		}

		return nil
	}
}
