package mappers

import (
	ent "barbershop-bot/entities"
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
)

type StorageToEntityMapper struct{}

var MapToEntity StorageToEntityMapper

func (s StorageToEntityMapper) Barber(barber st.Barber) (ent.Barber, error) {
	var br ent.Barber
	br.ID = barber.ID
	br.Name = mapBarberNameToEntity(barber.Name)
	br.Phone = mapBarberPhoneToEntity(barber.Phone)

	//st.Barber.LastWorkDate is always valid because default is not null.
	lastWorkDate, err := s.Date(barber.LastWorkDate)
	if err != nil {
		return ent.Barber{}, e.Wrap("can't map barber's last workdate to entity", err)
	}
	br.LastWorkdate = lastWorkDate
	return br, nil
}

func (s StorageToEntityMapper) Date(date string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, date, cfg.Location)
}

func (s StorageToEntityMapper) Service(service st.Service) (ent.Service, error) {
	duration, err := tm.ParseDuration(service.Duration)
	if err != nil {
		return ent.Service{}, e.Wrap("can't map duration to entity", err)
	}
	return ent.Service{
		ID:         service.ID,
		BarberID:   service.BarberID,
		Name:       service.Name,
		Desciption: service.Desciption,
		Price:      ent.Price(service.Price),
		Duration:   duration,
	}, nil
}

func (s StorageToEntityMapper) Workday(workday st.Workday) (ent.Workday, error) {
	date, err := s.Date(workday.Date)
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
