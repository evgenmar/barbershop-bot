package repository

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"errors"
	"time"
)

func mapDateToEntity(date string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, date, cfg.Location)
}

func mapBarberToEntity(barber *st.Barber) (ent.Barber, error) {
	var br ent.Barber
	if !barber.ID.Valid {
		return ent.Barber{}, errors.New("invalid barber ID")
	}
	br.ID = barber.ID.Int64
	br.Name = mapBarberNameToEntity(barber.Name)
	br.Phone = mapBarberPhoneToEntity(barber.Phone)
	status, err := mapStatusToEntity(&barber.Status)
	br.Status = status
	if err != nil {
		return br, e.Wrap("can't map Barber to entity", err)
	}
	return br, nil
}

func mapBarberNameToEntity(name sql.NullString) string {
	if !name.Valid {
		return ent.NoNameBarber
	} else {
		return name.String
	}
}

func mapBarberPhoneToEntity(phone sql.NullString) string {
	if !phone.Valid {
		return ent.NoPhoneBarber
	} else {
		return phone.String
	}
}

func mapStatusToEntity(status *st.Status) (ent.Status, error) {
	if !status.State.Valid || !status.Expiration.Valid {
		return ent.StatusStart, nil
	}
	expiration, err := time.Parse(time.DateTime, status.Expiration.String)
	if err != nil {
		return ent.StatusStart, e.Wrap("can't map Status to entity", err)
	}
	return ent.Status{
		State:      ent.State(status.State.Byte),
		Expiration: expiration,
	}, nil
}
