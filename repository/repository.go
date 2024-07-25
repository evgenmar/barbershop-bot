package repository

import (
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"
	tm "barbershop-bot/lib/time"
	m "barbershop-bot/repository/mappers"
	st "barbershop-bot/repository/storage"
	"context"
	"errors"
	"time"
)

type Repository struct {
	st.Storage
}

var (
	ErrAlreadyExists      = st.ErrAlreadyExists
	ErrAppointmentsExists = st.ErrAppointmentsExists
	ErrInvalidAppointment = errors.New("invalid appointment")
	ErrInvalidBarber      = errors.New("invalid barber")
	ErrInvalidDateRange   = errors.New("invalid date range")
	ErrInvalidService     = errors.New("invalid service")
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidWorkday     = errors.New("invalid workday")
	ErrNonUniqueData      = st.ErrNonUniqueData
	ErrNoSavedAppointment = errors.New("no saved appointment")
	ErrNoSavedBarber      = errors.New("no saved barber with specified ID")
	ErrNoSavedService     = errors.New("no saved service with specified ID")
	ErrNoSavedUser        = errors.New("no saved user with specified ID")
	ErrNoSavedWorkday     = errors.New("no saved workday with specified ID")
)

func New(storage st.Storage) Repository {
	return Repository{
		Storage: storage,
	}
}

func (r Repository) CreateAppointment(ctx context.Context, appt ent.Appointment) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	appointment, err := m.MapToStorage.AppointmentForCreate(appt)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidAppointment
		}
		return err
	}
	return r.Storage.CreateAppointment(ctx, appointment)
}

func (r Repository) CreateBarber(ctx context.Context, barberID int64) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	return r.Storage.CreateBarber(ctx, barberID)
}

func (r Repository) CreateService(ctx context.Context, service ent.Service) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	serv, err := m.MapToStorage.ServiceForCreate(service)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidService
		}
		return err
	}
	return r.Storage.CreateService(ctx, serv)
}

func (r Repository) CreateUser(ctx context.Context, user ent.User) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	ur, err := m.MapToStorage.User(user)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidUser
		}
		return err
	}
	return r.Storage.CreateUser(ctx, ur)
}

func (r Repository) CreateWorkdays(ctx context.Context, wds ...ent.Workday) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAlreadyExists) {
			err = ErrAlreadyExists
		}
	}()
	var workdays []st.Workday
	for _, workday := range wds {
		wd, err := m.MapToStorage.WorkdayForCreate(workday)
		if err != nil {
			if errors.Is(err, m.ErrInvalidEntity) {
				return ErrInvalidWorkday
			}
			return err
		}
		workdays = append(workdays, wd)
	}
	return r.Storage.CreateWorkdays(ctx, workdays...)
}

func (r Repository) DeleteAppointmentByID(ctx context.Context, appointmentID int) error {
	return r.Storage.DeleteAppointmentByID(ctx, appointmentID)
}

func (r Repository) DeleteBarberByID(ctx context.Context, barberID int64) error {
	return r.Storage.DeleteBarberByID(ctx, barberID)
}

func (r Repository) DeletePastAppointments(ctx context.Context, barberID int64) error {
	return r.Storage.DeletePastAppointments(ctx, barberID)
}

func (r Repository) DeleteServiceByID(ctx context.Context, serviceID int) error {
	return r.Storage.DeleteServiceByID(ctx, serviceID)
}

func (r Repository) DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange ent.DateRange) (err error) {
	defer func() {
		if errors.Is(err, st.ErrAppointmentsExists) {
			err = ErrAppointmentsExists
		}
	}()
	dr, err := m.MapToStorage.DateRange(dateRange)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidDateRange
		}
		return err
	}
	return r.Storage.DeleteWorkdaysByDateRange(ctx, barberID, dr)
}

func (r Repository) GetAllBarbers(ctx context.Context) (barbers []ent.Barber, err error) {
	brs, err := r.Storage.GetAllBarbers(ctx)
	if err != nil {
		return nil, err
	}
	for _, br := range brs {
		barber, err := m.MapToEntity.Barber(br)
		if err != nil {
			return nil, err
		}
		barbers = append(barbers, barber)
	}
	return barbers, nil
}

func (r Repository) GetAppointmentByID(ctx context.Context, appointmentID int) (appointment ent.Appointment, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedAppointment
		}
	}()
	appt, err := r.Storage.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return ent.Appointment{}, err
	}
	return m.MapToEntity.Appointment(appt), nil
}

func (r Repository) GetAppointmentIDByWorkdayIDAndTime(ctx context.Context, workdayID int, time tm.Duration) (appointmentID int, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedAppointment
		}
	}()
	return r.Storage.GetAppointmentIDByWorkdayIDAndTime(ctx, workdayID, m.MapToStorage.Duration(time))
}

func (r Repository) GetAppointmentsByDateRange(ctx context.Context, barberID int64, dateRange ent.DateRange) (appointments []ent.Appointment, err error) {
	defer func() { err = e.WrapIfErr("can't get appointments", err) }()
	dr, err := m.MapToStorage.DateRange(dateRange)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return nil, ErrInvalidDateRange
		}
		return nil, err
	}
	appts, err := r.Storage.GetAppointmentsByDateRange(ctx, barberID, dr)
	if err != nil {
		return nil, err
	}
	for _, appt := range appts {
		appointments = append(appointments, m.MapToEntity.Appointment(appt))
	}
	return appointments, nil
}

func (r Repository) GetBarberByID(ctx context.Context, barberID int64) (barber ent.Barber, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedBarber
		}
	}()
	br, err := r.Storage.GetBarberByID(ctx, barberID)
	if err != nil {
		return ent.Barber{}, err
	}
	return m.MapToEntity.Barber(br)
}

func (r Repository) GetLatestAppointmentDate(ctx context.Context, barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest appointment date", err) }()
	latestAD, err := r.Storage.GetLatestAppointmentDate(ctx, barberID)
	if err != nil {
		return time.Time{}, err
	}
	return m.MapToEntity.Date(latestAD)
}

func (r Repository) GetLatestWorkDate(ctx context.Context, barberID int64) (date time.Time, err error) {
	defer func() { err = e.WrapIfErr("can't get latest work date", err) }()
	latestWD, err := r.Storage.GetLatestWorkDate(ctx, barberID)
	if err != nil {
		return time.Time{}, err
	}
	return m.MapToEntity.Date(latestWD)
}

func (r Repository) GetServiceByID(ctx context.Context, serviceID int) (service ent.Service, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedService
		}
	}()
	serv, err := r.Storage.GetServiceByID(ctx, serviceID)
	if err != nil {
		return ent.Service{}, err
	}
	return m.MapToEntity.Service(serv), nil
}

func (r Repository) GetServicesByBarberID(ctx context.Context, barberID int64) (services []ent.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get services", err) }()
	servs, err := r.Storage.GetServicesByBarberID(ctx, barberID)
	if err != nil {
		return nil, err
	}
	for _, serv := range servs {
		services = append(services, m.MapToEntity.Service(serv))
	}
	return services, nil
}

func (r Repository) GetUpcomingAppointment(ctx context.Context, userID int64) (appointment ent.Appointment, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedAppointment
		}
	}()
	appt, err := r.Storage.GetUpcomingAppointment(ctx, userID)
	if err != nil {
		return ent.Appointment{}, err
	}
	return m.MapToEntity.Appointment(appt), nil
}

func (r Repository) GetUserByID(ctx context.Context, userID int64) (user ent.User, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedUser
		}
	}()
	ur, err := r.Storage.GetUserByID(ctx, userID)
	if err != nil {
		return ent.User{}, err
	}
	return m.MapToEntity.User(ur), nil
}

func (r Repository) GetWorkdayByID(ctx context.Context, workdayID int) (workday ent.Workday, err error) {
	defer func() {
		if errors.Is(err, st.ErrNoSavedObject) {
			err = ErrNoSavedWorkday
		}
	}()
	wd, err := r.Storage.GetWorkdayByID(ctx, workdayID)
	if err != nil {
		return ent.Workday{}, err
	}
	return m.MapToEntity.Workday(wd)
}

func (r Repository) GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange ent.DateRange) (workdays []ent.Workday, err error) {
	defer func() { err = e.WrapIfErr("can't get workdays", err) }()
	dr, err := m.MapToStorage.DateRange(dateRange)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return nil, ErrInvalidDateRange
		}
		return nil, err
	}
	wds, err := r.Storage.GetWorkdaysByDateRange(ctx, barberID, dr)
	if err != nil {
		return nil, err
	}
	for _, wd := range wds {
		workday, err := m.MapToEntity.Workday(wd)
		if err != nil {
			return nil, err
		}
		workdays = append(workdays, workday)
	}
	return workdays, nil
}

// UpdateAppointment updates only non-empty fields of Appointment. UserID, ServiceID, Duration, CreatedAt fields never updates.
func (r Repository) UpdateAppointment(ctx context.Context, appointment ent.Appointment) (err error) {
	defer func() {
		if errors.Is(err, st.ErrNonUniqueData) {
			err = ErrNonUniqueData
		}
	}()
	appt, err := m.MapToStorage.AppointmentForUpdate(appointment)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidAppointment
		}
		return err
	}
	return r.Storage.UpdateAppointment(ctx, appt)
}

// UpdateBarber updates only non-empty fields of Barber
func (r Repository) UpdateBarber(ctx context.Context, barber ent.Barber) (err error) {
	defer func() {
		if errors.Is(err, st.ErrNonUniqueData) {
			err = ErrNonUniqueData
		}
	}()
	br, err := m.MapToStorage.Barber(barber)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidBarber
		}
		return err
	}
	return r.Storage.UpdateBarber(ctx, br)
}

// UpdateService updates only non-empty fields of Service. BarberID field never updates.
func (r Repository) UpdateService(ctx context.Context, service ent.Service) (err error) {
	defer func() {
		if errors.Is(err, st.ErrNonUniqueData) {
			err = ErrNonUniqueData
		}
	}()
	serv, err := m.MapToStorage.ServiceForUpdate(service)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidService
		}
		return err
	}
	return r.Storage.UpdateService(ctx, serv)
}

// UpdateUser updates only non-empty fields of User
func (r Repository) UpdateUser(ctx context.Context, user ent.User) (err error) {
	ur, err := m.MapToStorage.User(user)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidUser
		}
		return err
	}
	return r.Storage.UpdateUser(ctx, ur)
}

// UpdateWorkday updates only non-empty fields of Workday. BarberID and Date fields never updates.
func (r Repository) UpdateWorkday(ctx context.Context, workday ent.Workday) (err error) {
	wd, err := m.MapToStorage.WorkdayForUpdate(workday)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			return ErrInvalidWorkday
		}
		return err
	}
	return r.Storage.UpdateWorkday(ctx, wd)
}
