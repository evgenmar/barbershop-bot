package repository

import (
	ent "barbershop-bot/entities"
	st "barbershop-bot/repository/storage"
	"regexp"
	"unicode"
)

type validatedEntityToStorageMapper struct {
	entityToStorageMapper
}

var mapToStorage validatedEntityToStorageMapper

func (v validatedEntityToStorageMapper) barber(barber ent.Barber) (st.Barber, error) {
	if barber.Name != "" && !isValidName(barber.Name) {
		return st.Barber{}, ErrInvalidName
	}
	if barber.Phone != "" {
		if !isValidPhone(barber.Phone) {
			return st.Barber{}, ErrInvalidPhone
		}
		barber.Phone = normalizePhone(barber.Phone)
	}
	return v.entityToStorageMapper.barber(barber)
}

func isValidName(text string) bool {
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

func isValidPhone(text string) bool {
	phonePattern := `^((\+7|8)[\s-]?)\(?\d{3}\)?[\s-]?\d{3}[\s-]?\d{2}[\s-]?\d{2}$`
	regex := regexp.MustCompile(phonePattern)
	return regex.MatchString(text)
}

func normalizePhone(phone string) (normalized string) {
	for _, r := range phone {
		if unicode.IsDigit(r) {
			normalized = normalized + string(r)
		}
	}
	return "+7" + normalized[1:]
}
