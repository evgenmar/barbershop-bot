package repository

import (
	"barbershop-bot/entities"
	"barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	"barbershop-bot/repository/storage"
	"database/sql"
	"errors"
	"time"
)

func mapDateToEntity(date string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, date, config.Location)
}

func mapBarberToEntity(barber *storage.Barber) (entities.Barber, error) {
	var br entities.Barber
	if !barber.ID.Valid {
		return entities.Barber{}, errors.New("invalid barber ID")
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
		return entities.NoPhoneBarber
	} else {
		return name.String
	}
}

func mapBarberPhoneToEntity(phone sql.NullString) string {
	if !phone.Valid {
		return entities.NoNameBarber
	} else {
		return phone.String
	}
}

func mapStatusToEntity(status *storage.Status) (entities.Status, error) {
	if !status.State.Valid || !status.Expiration.Valid {
		return entities.StatusStart, nil
	}
	expiration, err := time.Parse(time.DateTime, status.Expiration.String)
	if err != nil {
		return entities.StatusStart, e.Wrap("can't map Status to entity", err)
	}
	return entities.Status{
		State:      entities.State(status.State.Byte),
		Expiration: expiration,
	}, nil
}
