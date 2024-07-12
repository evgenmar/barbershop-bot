package repository

import (
	ent "barbershop-bot/entities"
	"barbershop-bot/lib/e"
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
	ErrInvalidBarber      = errors.New("invalid barber")
	ErrInvalidDateRange   = errors.New("invalid date range")
	ErrInvalidService     = errors.New("invalid service")
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidWorkday     = errors.New("invalid workday")
	ErrNonUniqueData      = st.ErrNonUniqueData
	ErrNoSavedBarber      = errors.New("no saved barber with specified ID")
	ErrNoSavedService     = errors.New("no saved service with specified ID")
	ErrNoSavedUser        = errors.New("no saved user with specified ID")
)

func New(storage st.Storage) Repository {
	return Repository{
		Storage: storage,
	}
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
	serv, err := m.MapToStorage.NewService(service)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			err = ErrInvalidService
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
			err = ErrInvalidUser
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
		wd, err := m.MapToStorage.Workday(workday)
		if err != nil {
			if errors.Is(err, m.ErrInvalidEntity) {
				err = ErrInvalidWorkday
			}
			return err
		}
		workdays = append(workdays, wd)
	}
	return r.Storage.CreateWorkdays(ctx, workdays...)
}

func (r Repository) DeleteAppointmentsBeforeDate(ctx context.Context, barberID int64, date time.Time) error {
	return r.Storage.DeleteAppointmentsBeforeDate(ctx, barberID, m.MapToStorage.Date(date))
}

func (r Repository) DeleteBarberByID(ctx context.Context, barberID int64) error {
	return r.Storage.DeleteBarberByID(ctx, barberID)
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
			err = ErrInvalidDateRange
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
	return m.MapToEntity.Service(serv)
}

func (r Repository) GetServicesByBarberID(ctx context.Context, barberID int64) (services []ent.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get services", err) }()
	servs, err := r.Storage.GetServicesByBarberID(ctx, barberID)
	if err != nil {
		return nil, err
	}
	for _, serv := range servs {
		service, err := m.MapToEntity.Service(serv)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
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

func (r Repository) GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange ent.DateRange) (workdays []ent.Workday, err error) {
	defer func() { err = e.WrapIfErr("can't get workdays", err) }()
	dr, err := m.MapToStorage.DateRange(dateRange)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			err = ErrInvalidDateRange
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
			err = ErrInvalidBarber
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
	serv, err := m.MapToStorage.UpdService(service)
	if err != nil {
		if errors.Is(err, m.ErrInvalidEntity) {
			err = ErrInvalidService
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
			err = ErrInvalidUser
		}
		return err
	}
	return r.Storage.UpdateUser(ctx, ur)
}
