package repository

import (
	ent "barbershop-bot/entities"
	tm "barbershop-bot/lib/time"
	st "barbershop-bot/repository/storage"
	"database/sql"
	"time"
)

func mapWorkdayToStorage(workday *ent.Workday) st.Workday {
	return st.Workday{
		BarberID:  mapIDToStorage(workday.BarberID),
		Date:      mapDateToStorage(workday.Date),
		StartTime: mapDurationToStorage(workday.StartTime),
		EndTime:   mapDurationToStorage(workday.EndTime),
	}
}

func mapIDToStorage(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: true}
}

func mapDateToStorage(date time.Time) sql.NullString {
	return sql.NullString{String: date.Format(time.DateOnly), Valid: true}
}

func mapDurationToStorage(dur tm.Duration) sql.NullString {
	return sql.NullString{String: dur.String(), Valid: true}
}

func mapBarberToStorage(barber *ent.Barber) st.Barber {
	return st.Barber{
		ID:     mapIDToStorage(barber.ID),
		Name:   mapNameToStorage(barber.Name),
		Phone:  mapPhoneToStorage(barber.Phone),
		Status: mapStatusToStorage(&barber.Status),
	}
}

func mapNameToStorage(name string) sql.NullString {
	if name == "" || name == ent.NoNameBarber {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: name, Valid: true}
}

func mapPhoneToStorage(phone string) sql.NullString {
	if phone == "" || phone == ent.NoPhoneBarber {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: phone, Valid: true}
}

func mapStatusToStorage(status *ent.Status) st.Status {
	return st.Status{
		State:      mapStateToStorage(status.State),
		Expiration: mapExpirationToStorage(status.Expiration),
	}
}

func mapStateToStorage(state ent.State) sql.NullByte {
	if state == 0 {
		return sql.NullByte{Byte: 0, Valid: false}
	}
	return sql.NullByte{Byte: byte(state), Valid: true}
}

func mapExpirationToStorage(exp time.Time) sql.NullString {
	if exp.Equal(time.Time{}) {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: exp.Format(time.DateTime), Valid: true}
}
