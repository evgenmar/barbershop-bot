package repository

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
)

type storageToEntityMapper struct{}

var mapToEntity storageToEntityMapper

func (s storageToEntityMapper) barber(barber st.Barber) (ent.Barber, error) {
	var br ent.Barber
	br.ID = barber.ID
	br.Name = mapBarberNameToEntity(barber.Name)
	br.Phone = mapBarberPhoneToEntity(barber.Phone)

	//st.Barber.LastWorkDate is always valid because default is not null.
	lastWorkDate, err := s.date(barber.LastWorkDate)
	if err != nil {
		return ent.Barber{}, e.Wrap("can't map barber's last workdate to entity", err)
	}
	br.LastWorkdate = lastWorkDate

	status, err := mapStatusToEntity(barber.Status)
	if err != nil {
		return ent.Barber{}, e.Wrap("can't map barber's status to entity", err)
	}
	br.Status = status
	return br, nil
}

func (s storageToEntityMapper) date(date string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, date, cfg.Location)
}

func (s storageToEntityMapper) workday(workday st.Workday) (ent.Workday, error) {
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
		ID:        workday.ID,
		BarberID:  workday.BarberID,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

func mapBarberNameToEntity(name sql.NullString) string {
	if !name.Valid {
		return ent.NoNameBarber
	}
	return name.String
}

func mapBarberPhoneToEntity(phone sql.NullString) string {
	if !phone.Valid {
		return ent.NoPhoneBarber
	}
	return phone.String
}

func mapStatusToEntity(stat st.Status) (ent.Status, error) {
	if !stat.State.Valid || !stat.Expiration.Valid {
		return ent.StatusStart, nil
	}
	expiration, err := time.Parse(time.DateTime, stat.Expiration.String)
	if err != nil {
		return ent.StatusStart, e.Wrap("can't map status to entity", err)
	}
	return ent.Status{
		State:      ent.State(stat.State.Byte),
		Expiration: expiration,
	}, nil
}
