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
	ErrInvalidEntity = errors.New("invalid entity")
)

func (v ValidatedEntityToStorageMapper) AppointmentForCreate(appt ent.Appointment) (st.Appointment, error) {
	if appt.UserID < 0 || appt.WorkdayID < 1 || appt.ServiceID < 1 || appt.Time < 0 || appt.Duration < 1 || appt.CreatedAt < 1 {
		return st.Appointment{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Appointment(appt), nil
}

func (v ValidatedEntityToStorageMapper) AppointmentForUpdate(appt ent.Appointment) (st.Appointment, error) {
	if appt.ID < 1 || appt.WorkdayID < 0 || appt.Time < 0 {
		return st.Appointment{}, ErrInvalidEntity
	}
	if appt.Note != "" && !IsValidNote(appt.Note) {
		return st.Appointment{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Appointment(appt), nil
}

func (v ValidatedEntityToStorageMapper) Barber(barber ent.Barber) (st.Barber, error) {
	if barber.ID == 0 {
		return st.Barber{}, ErrInvalidEntity
	}
	if barber.Name != "" && !isValidName(barber.Name) {
		return st.Barber{}, ErrInvalidEntity
	}
	if barber.Phone != "" && !isValidPhone(barber.Phone) {
		return st.Barber{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Barber(barber), nil
}

func (v ValidatedEntityToStorageMapper) Date(date time.Time) string {
	return v.EntityToStorageMapper.Date(date)
}

func (v ValidatedEntityToStorageMapper) DateRange(dateRange ent.DateRange) (st.DateRange, error) {
	if dateRange.FirstDate.After(dateRange.LastDate) {
		return st.DateRange{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.DateRange(dateRange), nil
}

func (v ValidatedEntityToStorageMapper) ServiceForCreate(service ent.Service) (st.Service, error) {
	if service.BarberID < 1 || service.Name == "" || service.Desciption == "" || service.Duration < 1 {
		return st.Service{}, ErrInvalidEntity
	}
	if service.Name != "" && !IsValidServiceName(service.Name) {
		return st.Service{}, ErrInvalidEntity
	}
	if service.Desciption != "" && !IsValidDescription(service.Desciption) {
		return st.Service{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Service(service), nil
}

func (v ValidatedEntityToStorageMapper) ServiceForUpdate(service ent.Service) (st.Service, error) {
	if service.ID < 1 {
		return st.Service{}, ErrInvalidEntity
	}
	if service.Name != "" && !IsValidServiceName(service.Name) {
		return st.Service{}, ErrInvalidEntity
	}
	if service.Desciption != "" && !IsValidDescription(service.Desciption) {
		return st.Service{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Service(service), nil
}

func (v ValidatedEntityToStorageMapper) User(user ent.User) (st.User, error) {
	if user.ID == 0 {
		return st.User{}, ErrInvalidEntity
	}
	if user.Name != "" && !isValidName(user.Name) {
		return st.User{}, ErrInvalidEntity
	}
	if user.Phone != "" && !isValidPhone(user.Phone) {
		return st.User{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.User(user), nil
}

func (v ValidatedEntityToStorageMapper) WorkdayForCreate(workday ent.Workday) (st.Workday, error) {
	if workday.BarberID == 0 || workday.Date.Equal(time.Time{}) || workday.StartTime < 0 || workday.EndTime <= workday.StartTime {
		return st.Workday{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Workday(workday), nil
}

func (v ValidatedEntityToStorageMapper) WorkdayForUpdate(workday ent.Workday) (st.Workday, error) {
	if workday.ID < 1 {
		return st.Workday{}, ErrInvalidEntity
	}
	return v.EntityToStorageMapper.Workday(workday), nil
}

func IsValidDescription(text string) bool {
	regex := regexp.MustCompile(`^[\p{L}\p{M}\p{N}\p{P}\p{Z}+\-]{10,400}$`)
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
	regex := regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9\s]{2,20}$`)
	var hasLetter bool
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	return regex.MatchString(text) && hasLetter && text != ent.NoName
}

func IsValidNote(text string) bool {
	regex := regexp.MustCompile(`^[\p{L}\p{M}\p{N}\p{P}\p{Z}+\-]{3,100}$`)
	return regex.MatchString(text)
}

func isValidPhone(text string) bool {
	regex := regexp.MustCompile(`^((\+7|8)[\s-]?)\(?\d{3}\)?[\s-]?\d{3}[\s-]?\d{2}[\s-]?\d{2}$`)
	return regex.MatchString(text)
}

func IsValidServiceName(text string) bool {
	regex := regexp.MustCompile(`^[\p{L}\p{M}\p{N}\p{P}\p{Z}+\-]{3,35}$`)
	var hasLetter bool
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	return regex.MatchString(text) && hasLetter && text != ent.NoName
}
