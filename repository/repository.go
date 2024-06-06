package repository

import (
	"barbershop-bot/entities"
	"barbershop-bot/lib/e"
	"barbershop-bot/repository/storage"
	"context"
	"errors"
	"time"
)

type Repository struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) CreateBarber(ctx context.Context, barberID int64) error {
	return r.storage.CreateBarber(ctx, barberID)
}

func (r *Repository) CreateWorkdays(ctx context.Context, wds ...entities.Workday) error {
	var workdays []storage.Workday
	for _, workday := range wds {
		workdays = append(workdays, mapWorkdayToStorage(&workday))
	}
	return r.storage.CreateWorkdays(ctx, workdays...)
}

func (r *Repository) FindAllBarberIDs(ctx context.Context) ([]int64, error) {
	return r.storage.FindAllBarberIDs(ctx)
}

func (r *Repository) GetBarberByID(ctx context.Context, barberID int64) (entities.Barber, error) {
	barber, err := r.storage.GetBarberByID(ctx, barberID)
	if err != nil {
		return entities.Barber{}, err
	}
	return mapBarberToEntity(&barber)
}

func (r *Repository) GetLatestWorkDate(ctx context.Context, barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	latestWD, err := r.storage.GetLatestWorkDate(ctx, barberID)
	if err != nil && !errors.Is(err, storage.ErrNoSavedWorkdates) {
		return time.Time{}, err
	}
	return mapDateToEntity(latestWD)
}

func (r *Repository) UpdateBarberName(ctx context.Context, name string, barberID int64) error {
	return r.storage.UpdateBarberName(ctx, name, barberID)
}

func (r *Repository) UpdateBarberPhone(ctx context.Context, phone string, barberID int64) error {
	return r.storage.UpdateBarberPhone(ctx, phone, barberID)
}

func (r *Repository) UpdateBarberStatus(ctx context.Context, status entities.Status, barberID int64) error {
	return r.storage.UpdateBarberStatus(ctx, mapStatusToStorage(&status), barberID)
}
