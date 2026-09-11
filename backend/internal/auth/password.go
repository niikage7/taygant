// Package auth отвечает за хеширование паролей и выпуск/разбор JWT.
// Пакет не знает про HTTP и про базу — только криптография и правила.
package auth

import (
	"errors"
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// minPasswordLength — нижняя граница длины пароля.
	minPasswordLength = 8
	// maxPasswordLength — ограничение самого bcrypt: он работает максимум
	// с 72 байтами и на более длинном вводе возвращает ошибку. Отсекаем сами,
	// чтобы пользователь получил внятное сообщение, а не 500.
	maxPasswordLength = 72
)

// ErrPasswordTooLong возвращается, когда пароль не помещается в ограничение bcrypt.
var ErrPasswordTooLong = fmt.Errorf("пароль длиннее %d байт", maxPasswordLength)

// ValidatePassword проверяет качество пароля при регистрации.
//
// Требования намеренно скромные: длина и наличие букв с цифрами. Более жёсткие
// правила (спецсимволы, проверка по словарям утечек) отсекают больше живых
// пользователей, чем реальных атак, и в MVP не оправданы.
func ValidatePassword(password string) error {
	if len([]rune(password)) < minPasswordLength {
		return fmt.Errorf("пароль должен быть не короче %d символов", minPasswordLength)
	}
	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("пароль должен содержать хотя бы одну букву и одну цифру")
	}
	return nil
}

// HashPassword считает bcrypt-хеш пароля.
//
// bcrypt выбран вместо Argon2id сознательно: он есть в golang.org/x/crypto без
// подбора параметров памяти и сам хранит соль и стоимость внутри хеша, так что
// в БД достаточно одной текстовой колонки.
func HashPassword(password string) (string, error) {
	if len(password) > maxPasswordLength {
		return "", ErrPasswordTooLong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("хеширование пароля: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword сравнивает пароль с сохранённым хешем.
//
// Возвращает bool, а не error: вызывающему коду нельзя различать «нет такого
// пользователя» и «неверный пароль» в ответе, иначе форма входа превращается
// в способ перебирать существующие адреса.
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
