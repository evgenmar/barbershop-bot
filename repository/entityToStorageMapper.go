package repository

import (
	ent "barbershop-bot/entities"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
)

type entityToStorageMapper struct{}

func (e *entityToStorageMapper) barber(barber *ent.Barber) (st.Barber, error) {
	var br st.Barber
	if barber.ID == 0 {
		return br, ErrInvalidID
	}
	br.ID = barber.ID
	if barber.Name != "" {
		br.Name = sql.NullString{String: barber.Name, Valid: true}
	}
	if barber.Phone != "" {
		br.Phone = sql.NullString{String: barber.Phone, Valid: true}
	}
	if !barber.LastWorkdate.Equal(time.Time{}) {
		br.LastWorkDate = e.date(barber.LastWorkdate)
	}
	if barber.State != 0 {
		br.State = sql.NullByte{Byte: byte(barber.State), Valid: true}
	}
	if !barber.Expiration.Equal(time.Time{}) {
		br.Expiration = sql.NullString{String: barber.Expiration.Format(time.DateTime), Valid: true}
	}
	return br, nil
}

func (e *entityToStorageMapper) date(date time.Time) string {
	return date.Format(time.DateOnly)
}

func (e *entityToStorageMapper) workday(workday *ent.Workday) (st.Workday, error) {
	if workday.BarberID == 0 || workday.Date.Equal(time.Time{}) || workday.StartTime < 0 || workday.EndTime < 0 {
		return st.Workday{}, ErrInvalidWorkday
	}
	return st.Workday{
		BarberID:  workday.BarberID,
		Date:      e.date(workday.Date),
		StartTime: workday.StartTime.String(),
		EndTime:   workday.EndTime.String(),
	}, nil
}
