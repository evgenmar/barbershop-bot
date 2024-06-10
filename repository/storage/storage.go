package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNoSavedWorkdates = errors.New("no saved workdates")
	ErrNoSavedBarber    = errors.New("no saved barber with specified ID")
	ErrNonUniqueData    = errors.New("data to save must be unique")
	ErrAlreadyExists    = errors.New("the object being saved already exists")
)

type Storage interface {
	//CreateBarber saves new BarberID to storage.
	CreateBarber(ctx context.Context, barberID int64) error

	//CreateWorkdays saves new Workdays to storage.
	CreateWorkdays(ctx context.Context, workdays ...Workday) error

	//FindAllBarberIDs return a slice of IDs of all barbers.
	FindAllBarberIDs(ctx context.Context) ([]int64, error)

	//GetBarberByID returns status of dialog for barber with barberID.
	GetBarberByID(ctx context.Context, barberID int64) (Barber, error)

	//GetLatestWorkDate returns the latest work date saved for barber with barberID.
	//If there is no saved work dates it returns ("2000-01-01", ErrNoSavedWorkdates).
	GetLatestWorkDate(ctx context.Context, barberID int64) (string, error)

	// Init prepares the storage for use. It creates the necessary tables if not exists.
	Init(ctx context.Context) error

	//UpdateBarber updates valid fields of Barber. ID field must be valid.
	UpdateBarber(ctx context.Context, barber Barber) error
}

type Workday struct {
	BarberID sql.NullInt64 `db:"barber_id"`

	//Date in YYYY-MM-DD format in local time zone
	Date sql.NullString `db:"date"`

	//Beginning of the working day in HH:MM in local time zone
	StartTime sql.NullString `db:"start_time"`

	//End of the working day in HH:MM in local time zone
	EndTime sql.NullString `db:"end_time"`
}

type Barber struct {
	ID   sql.NullInt64  `db:"id"`
	Name sql.NullString `db:"name"`

	//Format of phone number is +71234567890
	Phone sql.NullString `db:"phone"`
	Status
}

type Status struct {
	State sql.NullByte `db:"state"`

	//State expiration in YYYY-MM-DD HH:MM:SS format in UTC.
	Expiration sql.NullString `db:"state_expiration"`
}
