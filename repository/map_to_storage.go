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

func mapStatusToStorage(status *ent.Status) st.Status {
	return st.Status{
		State:      mapStateToStorage(status.State),
		Expiration: mapExpirationToStorage(status.Expiration),
	}
}

func mapStateToStorage(state ent.State) sql.NullByte {
	return sql.NullByte{Byte: byte(state), Valid: true}
}

func mapExpirationToStorage(exp time.Time) sql.NullString {
	return sql.NullString{String: exp.Format(time.DateTime), Valid: true}
}
