package mappers

import (
	ent "barbershop-bot/entities"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
	"unicode"
)

type EntityToStorageMapper struct{}

func (e EntityToStorageMapper) Barber(barber ent.Barber) st.Barber {
	return st.Barber{
		ID:           barber.ID,
		Name:         nullString(barber.Name),
		Phone:        nullString(normalizePhone(barber.Phone)),
		LastWorkDate: nullDate(barber.LastWorkdate),
	}
}

func (e EntityToStorageMapper) Date(date time.Time) string {
	return date.Format(time.DateOnly)
}

func (e EntityToStorageMapper) DateRange(dateRange ent.DateRange) st.DateRange {
	return st.DateRange{
		StartDate: e.Date(dateRange.StartDate),
		EndDate:   e.Date(dateRange.EndDate),
	}
}

func (e EntityToStorageMapper) Service(service ent.Service) st.Service {
	return st.Service{
		ID:         service.ID,
		BarberID:   service.BarberID,
		Name:       service.Name,
		Desciption: service.Desciption,
		Price:      uint(service.Price),
		Duration:   int16(service.Duration),
	}
}

func (e EntityToStorageMapper) User(user ent.User) st.User {
	return st.User{
		ID:    user.ID,
		Name:  nullString(user.Name),
		Phone: nullString(normalizePhone(user.Phone)),
	}
}

func (e EntityToStorageMapper) Workday(workday ent.Workday) st.Workday {
	return st.Workday{
		BarberID:  workday.BarberID,
		Date:      e.Date(workday.Date),
		StartTime: int16(workday.StartTime),
		EndTime:   int16(workday.EndTime),
	}
}

func normalizePhone(phone string) (normalized string) {
	if phone == "" {
		return ""
	}
	for _, r := range phone {
		if unicode.IsDigit(r) {
			normalized = normalized + string(r)
		}
	}
	return "+7" + normalized[1:]
}

func nullDate(date time.Time) (nullDt string) {
	if !date.Equal(time.Time{}) {
		nullDt = date.Format(time.DateOnly)
	}
	return
}

func nullString(str string) (nullStr sql.NullString) {
	if str != "" {
		nullStr = sql.NullString{String: str, Valid: true}
	}
	return
}
