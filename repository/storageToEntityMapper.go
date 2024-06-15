package repository

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	st "barbershop-bot/repository/storage"
	"time"
)

type storageToEntityMapper struct{}

var mapToEntity storageToEntityMapper

func (s *storageToEntityMapper) barber(barber *st.Barber) (ent.Barber, error) {
	var br ent.Barber
	br.ID = barber.ID

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
	lastWorkDate, err := s.date(barber.LastWorkDate)
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

func (s *storageToEntityMapper) workday(workday *st.Workday) (ent.Workday, error) {
	date, err := s.date(workday.Date)
	if err != nil {
		return ent.Workday{}, e.Wrap("can't map date to entity", err)
	}
	startTime, err := tm.ParseDuration(workday.StartTime)
	if err != nil {
		return ent.Workday{}, e.Wrap("can't map start time to entity", err)
	}
	endTime, err := tm.ParseDuration(workday.EndTime)
	if err != nil {
		return ent.Workday{}, e.Wrap("can't map end time to entity", err)
	}
	return ent.Workday{
		BarberID:  workday.BarberID,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}
