package contextprovider

import (
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"

	//tm "barbershop-bot/lib/time"
	rep "barbershop-bot/repository"
	st "barbershop-bot/repository/storage"
	"context"
	"time"
)

const (
	timoutWrite time.Duration = 2 * time.Second
	timoutRead                = 1 * time.Second
)

type ContextProvider struct {
	repo rep.Repository
}

var RepoWithContext ContextProvider

func InitRepoWithContext(storage st.Storage) {
	RepoWithContext = NewContextProvider(rep.New(storage))
}

func NewContextProvider(repository rep.Repository) ContextProvider {
	return ContextProvider{repo: repository}
}

// func (c ContextProvider) CreateAppointment(appt ent.Appointment) (err error) {
// 	defer func() { err = e.WrapIfErr("can't save new appointment", err) }()
// 	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
// 	defer cancel()
// 	return c.repo.CreateAppointment(ctx, appt)
// }

func (c ContextProvider) CreateBarber(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't save new barber ID", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.CreateBarber(ctx, barberID)
}

func (c ContextProvider) CreateService(service ent.Service) (err error) {
	defer func() { err = e.WrapIfErr("can't save new service", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.CreateService(ctx, service)
}

func (c ContextProvider) CreateUser(user ent.User) (err error) {
	defer func() { err = e.WrapIfErr("can't save new user", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.CreateUser(ctx, user)
}

func (c ContextProvider) CreateWorkdays(wds ...ent.Workday) (err error) {
	defer func() { err = e.WrapIfErr("can't create workdays", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.CreateWorkdays(ctx, wds...)
}

// func (c ContextProvider) DeleteAppointmentByID(appointmentID int) (err error) {
// 	defer func() { err = e.WrapIfErr("can't delete appointment", err) }()
// 	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
// 	defer cancel()
// 	return c.repo.DeleteAppointmentByID(ctx, appointmentID)
// }

func (c ContextProvider) DeleteBarberByID(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't delete barber", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.DeleteBarberByID(ctx, barberID)
}

func (c ContextProvider) DeletePastAppointments(barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't delete appointments", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.DeletePastAppointments(ctx, barberID)
}

func (c ContextProvider) DeleteServiceByID(serviceID int) (err error) {
	defer func() { err = e.WrapIfErr("can't delete service", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.DeleteServiceByID(ctx, serviceID)
}

func (c ContextProvider) DeleteWorkdaysByDateRange(barberID int64, dateRangeToDelete ent.DateRange) (err error) {
	defer func() { err = e.WrapIfErr("can't delete workdays", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.DeleteWorkdaysByDateRange(ctx, barberID, dateRangeToDelete)
}

func (c ContextProvider) GetAllBarbers() (barbers []ent.Barber, err error) {
	defer func() { err = e.WrapIfErr("can't get barbers", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetAllBarbers(ctx)
}

// func (c ContextProvider) GetAppointmentsByDateRange(barberID int64, dateRange ent.DateRange) (appointments []ent.Appointment, err error) {
// 	defer func() { err = e.WrapIfErr("can't get appointments", err) }()
// 	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
// 	defer cancel()
// 	return c.repo.GetAppointmentsByDateRange(ctx, barberID, dateRange)
// }

func (c ContextProvider) GetBarberByID(barberID int64) (barber ent.Barber, err error) {
	defer func() { err = e.WrapIfErr("can't get barber", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetBarberByID(ctx, barberID)
}

func (c ContextProvider) GetLatestAppointmentDate(barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest appointment date", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetLatestAppointmentDate(ctx, barberID)
}

func (c ContextProvider) GetLatestWorkDate(barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetLatestWorkDate(ctx, barberID)
}

func (c ContextProvider) GetServiceByID(serviceID int) (ervice ent.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get service", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetServiceByID(ctx, serviceID)
}

func (c ContextProvider) GetServicesByBarberID(barberID int64) (services []ent.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get services", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetServicesByBarberID(ctx, barberID)
}

// func (c ContextProvider) GetUpcomingAppointment(userID int64) (appointment ent.Appointment, err error) {
// 	defer func() { err = e.WrapIfErr("can't get appointment", err) }()
// 	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
// 	defer cancel()
// 	return c.repo.GetUpcomingAppointment(ctx, userID)
// }

func (c ContextProvider) GetUserByID(userID int64) (user ent.User, err error) {
	defer func() { err = e.WrapIfErr("can't get user", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
	defer cancel()
	return c.repo.GetUserByID(ctx, userID)
}

// func (c ContextProvider) GetWorkdaysByDateRange(barberID int64, dateRange ent.DateRange) (workdays []ent.Workday, err error) {
// 	defer func() { err = e.WrapIfErr("can't get workdays", err) }()
// 	ctx, cancel := context.WithTimeout(context.Background(), timoutRead)
// 	defer cancel()
// 	return c.repo.GetWorkdaysByDateRange(ctx, barberID, dateRange)
// }

// func (c ContextProvider) UpdateAppointmentTime(appointmentID, workdayID int, time tm.Duration) (err error) {
// 	defer func() { err = e.WrapIfErr("can't update appointment", err) }()
// 	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
// 	defer cancel()
// 	return c.repo.UpdateAppointmentTime(ctx, appointmentID, workdayID, time)
// }

func (c ContextProvider) UpdateBarber(barber ent.Barber) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.UpdateBarber(ctx, barber)
}

func (c ContextProvider) UpdateService(service ent.Service) (err error) {
	defer func() { err = e.WrapIfErr("can't update service", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.UpdateService(ctx, service)
}

func (c ContextProvider) UpdateUser(user ent.User) (err error) {
	defer func() { err = e.WrapIfErr("can't update user", err) }()
	ctx, cancel := context.WithTimeout(context.Background(), timoutWrite)
	defer cancel()
	return c.repo.UpdateUser(ctx, user)
}
