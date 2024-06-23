package mappers

import (
	ent "barbershop-bot/entities"
	st "barbershop-bot/repository/storage"
	"errors"
	"regexp"
	"time"
	"unicode"
)

type ValidatedEntityToStorageMapper struct {
	EntityToStorageMapper
}

var MapToStorage ValidatedEntityToStorageMapper

var (
	ErrInvalidBarber    = errors.New("invalid barber")
	ErrInvalidDate      = errors.New("invalid date")
	ErrInvalidDateRange = errors.New("invalid date range")
	ErrInvalidService   = errors.New("invalid service")
	ErrInvalidWorkday   = errors.New("invalid workday")
)

func (v ValidatedEntityToStorageMapper) Barber(barber ent.Barber) (st.Barber, error) {
	if barber.ID == 0 {
		return st.Barber{}, ErrInvalidBarber
	}
	if barber.Name != "" && !isValidName(barber.Name) {
		return st.Barber{}, ErrInvalidBarber
	}
	if barber.Phone != "" {
		if !isValidPhone(barber.Phone) {
			return st.Barber{}, ErrInvalidBarber
		}
	}
	return v.EntityToStorageMapper.Barber(barber), nil
}

func (v ValidatedEntityToStorageMapper) Date(date time.Time) (string, error) {
	if date.Year() < 2000 {
		return "", ErrInvalidDate
	}
	return v.EntityToStorageMapper.Date(date), nil
}

func (v ValidatedEntityToStorageMapper) DateRange(dateRange ent.DateRange) (st.DateRange, error) {
	if dateRange.StartDate.After(dateRange.EndDate) {
		return st.DateRange{}, ErrInvalidDateRange
	}
	return v.EntityToStorageMapper.DateRange(dateRange), nil
}

func (v ValidatedEntityToStorageMapper) Service(service ent.Service) (st.Service, error) {
	if service.ID < 0 || service.BarberID < 0 || service.Price < 0 || service.Duration < 0 {
		return st.Service{}, ErrInvalidService
	}
	if service.Name != "" && !isValidName(service.Name) {
		return st.Service{}, ErrInvalidService
	}
	if service.Desciption != "" && !isValidDescription(service.Desciption) {
		return st.Service{}, ErrInvalidService
	}
	return v.EntityToStorageMapper.Service(service), nil
}

func (v ValidatedEntityToStorageMapper) Workday(workday ent.Workday) (st.Workday, error) {
	if workday.BarberID == 0 || workday.Date.Equal(time.Time{}) || workday.StartTime < 0 || workday.EndTime <= workday.StartTime {
		return st.Workday{}, ErrInvalidWorkday
	}
	return v.EntityToStorageMapper.Workday(workday), nil
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