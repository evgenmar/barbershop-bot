package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNoSavedWorkdates = errors.New("no saved workdates")
	ErrNonUniqueData    = errors.New("data to save must be unique")
)

type Storage interface {
	//CreateBarber saves new BarberID to storage
	CreateBarber(ctx context.Context, barberID int64) error

	//CreateWorkdays saves new Workdays to storage
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

	//IsBarberExists reports if barber with specified ID exists in storage
	IsBarberExists(ctx context.Context, barberID int64) (bool, error)

	// UpdateBarberNameAndStatus saves new name and status for barber with barberID.
	UpdateBarberNameAndStatus(ctx context.Context, name string, status Status, barberID int64) error

	// UpdateBarberPhoneAndStatus saves new phone and status for barber with barberID.
	UpdateBarberPhoneAndStatus(ctx context.Context, phone string, status Status, barberID int64) error

	// UpdateBarberStatus saves new status for barber with barberID.
	UpdateBarberStatus(ctx context.Context, status Status, barberID int64) error
}

type Workday struct {
	BarberID int64 `db:"barber_id"`

	//Date in YYYY-MM-DD format in local time zone
	Date string `db:"date"`

	//Beginning of the working day in HH:MM in local time zone
	StartTime string `db:"start_time"`

	//End of the working day in HH:MM in local time zone
	EndTime string `db:"end_time"`
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
