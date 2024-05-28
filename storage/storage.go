package storage

import (
	"context"
	"errors"
)

var ErrNoSavedWorkdates = errors.New("no saved workdates")

type Storage interface {
	Close() error

	//CreateBarberID saves new BarberID to storage
	CreateBarberID(ctx context.Context, barberID int64) error

	//CreateWorkdays saves new Workdays to storage
	CreateWorkdays(ctx context.Context, workdays ...Workday) error

	//FindAllBarberIDs return a slice of IDs of all barbers.
	FindAllBarberIDs(ctx context.Context) ([]int64, error)

	//GetBarberStatus returns status of dialog for barber with barberID.
	GetBarberStatus(ctx context.Context, barberID int64) (Status, error)

	//GetLatestWorkDate returns the latest work date saved for barber with barberID.
	//If there is no saved work dates it returns ("2000-01-01", ErrNoSavedWorkdates).
	GetLatestWorkDate(ctx context.Context, barberID int64) (string, error)

	// Init prepares the storage for use. It creates the necessary tables if not exists.
	Init(ctx context.Context) error

	// UpdateBarberNameAndStatus saves new name and status for barber with barberID.
	UpdateBarberNameAndStatus(ctx context.Context, name string, status Status, barberID int64) error

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

type Status struct {
	State uint8 `db:"state"`

	//State expiration in YYYY-MM-DD HH:MM:SS format in UTC.
	Expiration string `db:"state_expiration"`
}
