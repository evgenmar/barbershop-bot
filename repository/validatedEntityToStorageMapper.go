package repository

import (
	ent "barbershop-bot/entities"
	st "barbershop-bot/repository/storage"
	"regexp"
	"time"
	"unicode"
)

type validatedEntityToStorageMapper struct {
	entityToStorageMapper
}

var mapToStorage validatedEntityToStorageMapper

func (v validatedEntityToStorageMapper) barber(barber ent.Barber) (st.Barber, error) {
	if barber.ID == 0 {
		return st.Barber{}, ErrInvalidID
	}
	if barber.Name != "" && !isValidName(barber.Name) {
		return st.Barber{}, ErrInvalidName
	}
	if barber.Phone != "" {
		if !isValidPhone(barber.Phone) {
			return st.Barber{}, ErrInvalidPhone
		}
		barber.Phone = normalizePhone(barber.Phone)
	}
	return v.entityToStorageMapper.barber(barber), nil
}

func (v validatedEntityToStorageMapper) date(date time.Time) (string, error) {
	if date.Year() < 2000 {
		return "", ErrInvalidDate
	}
	return v.entityToStorageMapper.date(date), nil
}

func (v validatedEntityToStorageMapper) dateRange(dateRange ent.DateRange) (st.DateRange, error) {
	if dateRange.StartDate.After(dateRange.EndDate) {
		return st.DateRange{}, ErrInvalidDateRange
	}
	return v.entityToStorageMapper.dateRange(dateRange), nil
}

func (v validatedEntityToStorageMapper) service(service ent.Service) (st.Service, error) {
	if service.ID < 0 || service.BarberID < 0 || service.Price < 0 || service.Duration < 0 {
		return st.Service{}, ErrInvalidService
	}
	if service.Name != "" && !isValidName(service.Name) {
		return st.Service{}, ErrInvalidService
	}
	if service.Desciption != "" && !isValidDescription(service.Desciption) {
		return st.Service{}, ErrInvalidService
	}
	return v.entityToStorageMapper.service(service), nil
}

func (v validatedEntityToStorageMapper) workday(workday ent.Workday) (st.Workday, error) {
	if workday.BarberID == 0 || workday.Date.Equal(time.Time{}) || workday.StartTime < 0 || workday.EndTime <= workday.StartTime {
		return st.Workday{}, ErrInvalidWorkday
	}
	return v.entityToStorageMapper.workday(workday), nil
}

func isValidDescription(text string) bool {
	namePattern := `^[a-zA-Zа-яА-Я0-9,.\s]{10,400}$`
	regex := regexp.MustCompile(namePattern)
	var has7Letters bool
	nLetters := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			nLetters++
			if nLetters > 6 {
				has7Letters = true
				break
			}
		}
	}
	return regex.MatchString(text) && has7Letters
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
