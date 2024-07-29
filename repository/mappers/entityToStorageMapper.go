package mappers

import (
	ent "barbershop-bot/entities"
	tm "barbershop-bot/lib/time"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
	"unicode"

	tele "gopkg.in/telebot.v3"
)

type EntityToStorageMapper struct{}

func (e EntityToStorageMapper) Appointment(appointment ent.Appointment) st.Appointment {
	return st.Appointment{
		ID:        appointment.ID,
		UserID:    mapInt64ToNullInt64(appointment.UserID),
		WorkdayID: appointment.WorkdayID,
		ServiceID: mapIntToNullInt32(appointment.ServiceID),
		Time:      e.Duration(appointment.Time),
		Duration:  e.Duration(appointment.Duration),
		Note:      mapStrToNullStr(appointment.Note),
		CreatedAt: appointment.CreatedAt,
	}
}

func (e EntityToStorageMapper) Barber(barber ent.Barber) st.Barber {
	return st.Barber{
		ID:            barber.ID,
		Name:          mapStrToNullStr(barber.Name),
		Phone:         mapStrToNullStr(normalizePhone(barber.Phone)),
		LastWorkDate:  mapDateToStr(barber.LastWorkdate),
		StoredMessage: tele.StoredMessage{MessageID: barber.MessageID, ChatID: barber.ChatID},
	}
}

func (e EntityToStorageMapper) Date(date time.Time) string {
	return date.Format(time.DateOnly)
}

func (e EntityToStorageMapper) DateRange(dateRange ent.DateRange) st.DateRange {
	return st.DateRange{
		FirstDate: e.Date(dateRange.FirstDate),
		LastDate:  e.Date(dateRange.LastDate),
	}
}

func (e EntityToStorageMapper) Duration(dur tm.Duration) int16 {
	return int16(dur)
}

func (e EntityToStorageMapper) Service(service ent.Service) st.Service {
	return st.Service{
		ID:         service.ID,
		BarberID:   service.BarberID,
		Name:       service.Name,
		Desciption: service.Desciption,
		Price:      uint(service.Price),
		Duration:   e.Duration(service.Duration),
	}
}

func (e EntityToStorageMapper) User(user ent.User) st.User {
	return st.User{
		ID:            user.ID,
		Name:          mapStrToNullStr(user.Name),
		Phone:         mapStrToNullStr(normalizePhone(user.Phone)),
		StoredMessage: tele.StoredMessage{MessageID: user.MessageID, ChatID: user.ChatID},
	}
}

func (e EntityToStorageMapper) Workday(workday ent.Workday) st.Workday {
	return st.Workday{
		ID:        workday.ID,
		BarberID:  workday.BarberID,
		Date:      e.Date(workday.Date),
		StartTime: e.Duration(workday.StartTime),
		EndTime:   e.Duration(workday.EndTime),
	}
}

func normalizePhone(phone string) (normalized string) {
	if phone == "" {
		return ""
	}
	for _, r := range phone {
		if unicode.IsDigit(r) {
			normalized = normalized + string(r)
		}
	}
	return "+7" + normalized[1:]
}

func mapDateToStr(date time.Time) (str string) {
	if !date.Equal(time.Time{}) {
		str = date.Format(time.DateOnly)
	}
	return
}

func mapIntToNullInt32(i int) (nullInt32 sql.NullInt32) {
	if i != 0 {
		nullInt32 = sql.NullInt32{Int32: int32(i), Valid: true}
	}
	return
}

func mapInt64ToNullInt64(i int64) (nullInt64 sql.NullInt64) {
	if i != 0 {
		nullInt64 = sql.NullInt64{Int64: i, Valid: true}
	}
	return
}

func mapStrToNullStr(str string) (nullStr sql.NullString) {
	if str != "" {
		nullStr = sql.NullString{String: str, Valid: true}
	}
	return
}
