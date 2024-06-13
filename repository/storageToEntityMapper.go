package repository

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	st "barbershop-bot/repository/storage"
	"errors"
	"time"
)

type storageToEntityMapper struct{}

var mapToEntity storageToEntityMapper

func (s *storageToEntityMapper) barber(barber *st.Barber) (ent.Barber, error) {
	var br ent.Barber
	if !barber.ID.Valid {
		return ent.Barber{}, errors.New("invalid barber ID")
	}
	br.ID = barber.ID.Int64

	if !barber.Name.Valid {
		br.Name = ent.NoNameBarber
	} else {
		br.Name = barber.Name.String
	}

	if !barber.Phone.Valid {
		br.Phone = ent.NoPhoneBarber
	} else {
		br.Phone = barber.Phone.String
	}

	//st.Barber.LastWorkDate is always valid because default is not null.
	lastWorkDate, err := s.date(barber.LastWorkDate.String)
	if err != nil {
		return ent.Barber{}, e.Wrap("can't map barber's last workdate to entity", err)
	}
	br.LastWorkdate = lastWorkDate

	if !barber.State.Valid || !barber.Expiration.Valid {
		br.Status = ent.StatusStart
		return br, nil
	}
	expiration, err := time.Parse(time.DateTime, barber.Expiration.String)
	if err != nil {
		br.Status = ent.StatusStart
		return ent.Barber{}, e.Wrap("can't map barber's status to entity", err)
	}
	br.Status = ent.Status{
		State:      ent.State(barber.State.Byte),
		Expiration: expiration,
	}
	return br, nil
}

func (s *storageToEntityMapper) date(date string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, date, cfg.Location)
}
