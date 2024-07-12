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
	br.Name = mapNameToEntity(barber.Name)
	br.Phone = mapPhoneToEntity(barber.Phone)

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

func (s StorageToEntityMapper) Service(service st.Service) ent.Service {
	return ent.Service{
		ID:         service.ID,
		BarberID:   service.BarberID,
		Name:       service.Name,
		Desciption: service.Desciption,
		Price:      ent.Price(service.Price),
		Duration:   tm.Duration(service.Duration),
	}
}

func (s StorageToEntityMapper) User(user st.User) ent.User {
	return ent.User{
		ID:    user.ID,
		Name:  mapNameToEntity(user.Name),
		Phone: mapPhoneToEntity(user.Phone),
	}
}

func (s StorageToEntityMapper) Workday(workday st.Workday) (ent.Workday, error) {
	date, err := s.Date(workday.Date)
	if err != nil {
		return ent.Workday{}, e.Wrap("can't map workday to entity", err)
	}
	return ent.Workday{
		ID:        workday.ID,
		BarberID:  workday.BarberID,
		Date:      date,
		StartTime: tm.Duration(workday.StartTime),
		EndTime:   tm.Duration(workday.EndTime),
	}, nil
}

func mapNameToEntity(name sql.NullString) string {
	if !name.Valid {
		return ent.NoName
	}
	return name.String
}

func mapPhoneToEntity(phone sql.NullString) string {
	if !phone.Valid {
		return ent.NoPhone
	}
	return phone.String
}
