package utils

import (
	ent "barbershop-bot/entities"
	"regexp"
	"unicode"
)

func IsValidName(text string) bool {
	namePattern := `^[a-zA-Zа-яА-Я0-9\s]{2,20}$`
	regex := regexp.MustCompile(namePattern)
	var hasLetter bool
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	return regex.MatchString(text) && hasLetter && text != ent.NoNameBarber
}

func IsValidPhone(text string) bool {
	phonePattern := `^((\+7|8)[\s-]?)\(?\d{3}\)?[\s-]?\d{3}[\s-]?\d{2}[\s-]?\d{2}$`
	regex := regexp.MustCompile(phonePattern)
	return regex.MatchString(text)
}

func NormalizePhone(phone string) (normalized string) {
	for _, r := range phone {
		if unicode.IsDigit(r) {
			normalized = normalized + string(r)
		}
	}
	return "+7" + normalized[1:]
}
