package contextprovider

import (
	st "barbershop-bot/repository/storage"
	"context"
)

type StorageContextProvider struct {
	st.Storage
}

func NewStorageContextProvider(storage st.Storage) StorageContextProvider {
	return StorageContextProvider{Storage: storage}
}

func (s StorageContextProvider) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return s.Storage.Init(ctx)
}

func (s StorageContextProvider) FindAllBarberIDs() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return s.Storage.FindAllBarberIDs(ctx)
}

func (s StorageContextProvider) CreateBarber(barberID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return s.Storage.CreateBarber(ctx, barberID)
}
