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
	var br st.Barber
	br.ID = barber.ID
	br.Name = nullString(barber.Name)
	br.Phone = nullString(normalizePhone(barber.Phone))
	if !barber.LastWorkdate.Equal(time.Time{}) {
		br.LastWorkDate = e.Date(barber.LastWorkdate)
	}
	return br
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
		Duration:   service.Duration.ShortString(),
	}
}

func (e EntityToStorageMapper) Workday(workday ent.Workday) st.Workday {
	return st.Workday{
		BarberID:  workday.BarberID,
		Date:      e.Date(workday.Date),
		StartTime: workday.StartTime.ShortString(),
		EndTime:   workday.EndTime.ShortString(),
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

func nullString(str string) (nullStr sql.NullString) {
	if str != "" {
		nullStr = sql.NullString{String: str, Valid: true}
	}
	return
}
