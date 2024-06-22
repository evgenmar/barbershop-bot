package repository

import (
	ent "barbershop-bot/entities"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
)

type entityToStorageMapper struct{}

func (e entityToStorageMapper) barber(barber ent.Barber) st.Barber {
	var br st.Barber
	br.ID = barber.ID
	br.Name = nullString(barber.Name)
	br.Phone = nullString(barber.Phone)
	if !barber.LastWorkdate.Equal(time.Time{}) {
		br.LastWorkDate = e.date(barber.LastWorkdate)
	}
	br.Status = mapStatusToStorage(barber.Status)
	return br
}

func (e entityToStorageMapper) date(date time.Time) string {
	return date.Format(time.DateOnly)
}

func (e entityToStorageMapper) dateRange(dateRange ent.DateRange) st.DateRange {
	return st.DateRange{
		StartDate: e.date(dateRange.StartDate),
		EndDate:   e.date(dateRange.EndDate),
	}
}

func (e entityToStorageMapper) service(service ent.Service) st.Service {
	return st.Service{
		ID:         service.ID,
		BarberID:   service.BarberID,
		Name:       service.Name,
		Desciption: service.Desciption,
		Price:      service.Price,
		Duration:   service.Duration.String(),
	}
}

func (e entityToStorageMapper) workday(workday ent.Workday) st.Workday {
	return st.Workday{
		BarberID:  workday.BarberID,
		Date:      e.date(workday.Date),
		StartTime: workday.StartTime.String(),
		EndTime:   workday.EndTime.String(),
	}
}

func mapStatusToStorage(stat ent.Status) (status st.Status) {
	if stat.State != 0 {
		status.State = sql.NullByte{Byte: byte(stat.State), Valid: true}
	}
	if !stat.Expiration.Equal(time.Time{}) {
		status.Expiration = sql.NullString{String: stat.Expiration.Format(time.DateTime), Valid: true}
	}
	return
}

func nullString(str string) (nullStr sql.NullString) {
	if str != "" {
		nullStr = sql.NullString{String: str, Valid: true}
	}
	return
}
