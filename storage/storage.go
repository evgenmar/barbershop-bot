package storage

import (
	"context"
	"errors"
	"time"
)

var ErrNoSavedWorkdates = errors.New("no saved workdates")

type Storage interface {
	//Init prepares the storage for use. It creates the necessary tables if they do not exist
	//and imports the barber ID from the environment variable into the storage
	//if there is not a single barber in the storage.
	Init(ctx context.Context) error

	//BarberIDs return a slice of IDs of all barbers.
	BarberIDs(ctx context.Context) ([]int64, error)

	//LatestWorkDate return the latest work date saved for barber with barberID.
	//Result have a format of time.Time with HH:MM:SS set to 00:00:00.
	LatestWorkDate(ctx context.Context, barberID int64) (time.Time, error)

	//SaveWorkday saves Workday type value to storage
	SaveWorkday(ctx context.Context, workday Workday) error

	//SaveBarberName saves name for barber with barberID
	SaveBarberName(ctx context.Context, name string, barberID int64) error

	//SaveBarberState saves state of dialog for barber with barberID.
	//It also saves expiration time for this state taking it as 24 hours after SaveBarberState call.
	SaveBarberState(ctx context.Context, state uint8, barberID int64) error

	Close() error
}

type Workday struct {
	BarberID  int64
	Date      time.Time
	StartTime string
	EndTime   string
}
