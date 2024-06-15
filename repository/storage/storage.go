package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNoSavedWorkdates   = errors.New("no saved workdates")
	ErrNoSavedBarber      = errors.New("no saved barber with specified ID")
	ErrNonUniqueData      = errors.New("data to save must be unique")
	ErrAlreadyExists      = errors.New("the object being saved already exists")
	ErrAppointmentsExists = errors.New("there are active appointments for the period being deleted")
)

type Storage interface {
	//CreateBarber saves new BarberID to storage.
	CreateBarber(ctx context.Context, barberID int64) error

	//CreateWorkdays saves new Workdays to storage.
	CreateWorkdays(ctx context.Context, workdays ...Workday) error

	//DeleteWorkdaysByDateRange removes working days that fall within the date range.
	//It only removes working days for barber with specified barberID.
	DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange DateRange) error

	//FindAllBarberIDs return a slice of IDs of all barbers.
	FindAllBarberIDs(ctx context.Context) ([]int64, error)

	//GetBarberByID returns barber with barberID.
	GetBarberByID(ctx context.Context, barberID int64) (Barber, error)

	//GetLatestWorkDate returns the latest work date saved for barber with barberID.
	//If there is no saved work dates it returns ("2000-01-01", ErrNoSavedWorkdates).
	GetLatestWorkDate(ctx context.Context, barberID int64) (string, error)

	// GetWorkdaysByDateRange returns working days that fall within the date range.
	// It only returns working days for barber with specified barberID.
	// Returned working days are sorted by date in ascending order.
	GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange DateRange) ([]Workday, error)

	// Init prepares the storage for use. It creates the necessary tables if not exists.
	Init(ctx context.Context) error

	//UpdateBarber updates valid fields of Barber. ID field must be valid.
	UpdateBarber(ctx context.Context, barber Barber) error
}

type Barber struct {
	ID   int64          `db:"id"`
	Name sql.NullString `db:"name"`

	//Format of phone number is +71234567890.
	Phone sql.NullString `db:"phone"`

	//LastWorkdate is a date in YYYY-MM-DD format in local time zone. Default is '3000-01-01'.
	LastWorkDate string `db:"last_workdate"`
	Status
}

type DateRange struct {
	StartDate string
	EndDate   string
}

type Status struct {
	State sql.NullByte `db:"state"`

	//State expiration in YYYY-MM-DD HH:MM:SS format in UTC.
	Expiration sql.NullString `db:"state_expiration"`
}

type Workday struct {
	BarberID int64 `db:"barber_id"`

	//Date in YYYY-MM-DD format in local time zone.
	Date string `db:"date"`

	//Beginning of the working day in HH:MM in local time zone.
	StartTime string `db:"start_time"`

	//End of the working day in HH:MM in local time zone.
	EndTime string `db:"end_time"`
}
