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
	br.ID = sql.NullInt64{Int64: barber.ID, Valid: true}

	if barber.Name != "" {
		br.Name = sql.NullString{String: barber.Name, Valid: true}
	}

	if barber.Phone != "" {
		br.Phone = sql.NullString{String: barber.Phone, Valid: true}
	}

	if !barber.LastWorkdate.Equal(time.Time{}) {
		br.LastWorkDate = sql.NullString{String: e.date(barber.LastWorkdate), Valid: true}
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
	if workday.BarberID == 0 || workday.Date.Equal(time.Time{}) || workday.StartTime <= 0 || workday.EndTime <= 0 {
		return st.Workday{}, ErrInvalidWorkday
	}
	return st.Workday{
		BarberID:  sql.NullInt64{Int64: workday.BarberID, Valid: true},
		Date:      sql.NullString{String: e.date(workday.Date), Valid: true},
		StartTime: sql.NullString{String: workday.StartTime.String(), Valid: true},
		EndTime:   sql.NullString{String: workday.EndTime.String(), Valid: true},
	}, nil
}