package contextprovider

import (
	"context"

	st "github.com/evgenmar/barbershop-bot/repository/storage"
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
