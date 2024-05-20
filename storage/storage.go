package storage

import "context"

type Storage interface {
	//Init prepares the storage for use. It creates the necessary tables if they do not exist
	//and imports the barber ID from the environment variable into the storage
	//if there is not a single barber in the storage.
	Init(ctx context.Context) error

	//BarberIDs return a slice of IDs of all barbers.
	BarberIDs(ctx context.Context) ([]int64, error)

	Close() error
}
