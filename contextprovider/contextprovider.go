package contextprovider

import (
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"
	rep "barbershop-bot/repository"
	"context"
	"time"
)

const (
	TimoutWrite time.Duration = 2 * time.Second
	TimoutRead                = 1 * time.Second
)

type ContextProvider struct{}

var CP ContextProvider

func (c ContextProvider) CreateBarber(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't save new barber ID", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutWrite)
	defer cancel()
	return rep.Rep.CreateBarber(ctx, barberID)
}

func (c ContextProvider) CreateWorkdays(wds ...ent.Workday) (err error) {
	defer func() { err = e.WrapIfErr("can't create workdays", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutWrite)
	defer cancel()
	return rep.Rep.CreateWorkdays(ctx, wds...)
}

func (c ContextProvider) DeleteWorkdaysByDateRange(barberID int64, dateRangeToDelete ent.DateRange) (err error) {
	defer func() { err = e.WrapIfErr("can't delete workdays", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutWrite)
	defer cancel()
	return rep.Rep.DeleteWorkdaysByDateRange(ctx, barberID, dateRangeToDelete)
}

func (c ContextProvider) FindAllBarberIDs() (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutRead)
	defer cancel()
	return rep.Rep.FindAllBarberIDs(ctx)
}

func (c ContextProvider) GetBarberByID(barberID int64) (barber ent.Barber, err error) {
	defer func() { err = e.WrapIfErr("can't get barber", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutRead)
	defer cancel()
	return rep.Rep.GetBarberByID(ctx, barberID)
}

func (c ContextProvider) GetLatestAppointmentDate(barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest appointment date", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutRead)
	defer cancel()
	return rep.Rep.GetLatestAppointmentDate(ctx, barberID)
}

func (c ContextProvider) GetLatestWorkDate(barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutRead)
	defer cancel()
	return rep.Rep.GetLatestWorkDate(ctx, barberID)
}

func (c ContextProvider) GetWorkdaysByDateRange(barberID int64, dateRange ent.DateRange) (workdays []ent.Workday, err error) {
	defer func() { err = e.WrapIfErr("can't get workdays", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutRead)
	defer cancel()
	return rep.Rep.GetWorkdaysByDateRange(ctx, barberID, dateRange)
}

func (c ContextProvider) UpdateBarber(barber ent.Barber) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), TimoutWrite)
	defer cancel()
	return rep.Rep.UpdateBarber(ctx, barber)
}
