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

	//CreateWorkday saves new Workday to storage
	CreateWorkday(ctx context.Context, workday Workday) error

	//FindAllBarberIDs return a slice of IDs of all barbers.
	FindAllBarberIDs(ctx context.Context) ([]int64, error)

	//GetLatestWorkDate returns the latest work date saved for barber with barberID.
	//If there is no saved work dates it returns ("2000-01-01", ErrNoSavedWorkdates).
	GetLatestWorkDate(ctx context.Context, barberID int64) (string, error)

	// Init prepares the storage for use. It creates the necessary tables if not exists.
	Init(ctx context.Context) error

	//UpdateBarberName saves new name for barber with barberID
	UpdateBarberName(ctx context.Context, name string, barberID int64) error

	//UpdateBarberStatus saves new status of dialog for barber with barberID.
	UpdateBarberStatus(ctx context.Context, status Status, barberID int64) error
}

type Workday struct {
	BarberID int64

	//Date in YYYY-MM-DD format in local time zone
	Date string

	//Beginning of the working day in HH:MM in local time zone
	StartTime string

	//End of the working day in HH:MM in local time zone
	EndTime string
}

type Status struct {
	State uint8

	//State expiration in YYYY-MM-DD HH:MM:SS format in UTC.
	Expiration string
}
