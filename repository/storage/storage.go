package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrAlreadyExists        = errors.New("the object being saved already exists")
	ErrAppointmentsExists   = errors.New("there are active appointments for the period being deleted")
	ErrInvalidService       = errors.New("invalid service")
	ErrNonUniqueData        = errors.New("data to save must be unique")
	ErrNoSavedBarber        = errors.New("no saved barber with specified ID")
	ErrUnspecifiedServiceID = errors.New("unspecified serviceID")
)

type Storage interface {
	//CreateBarber saves new BarberID to storage.
	CreateBarber(ctx context.Context, barberID int64) error

	//CreateService saves new service to storage.
	CreateService(ctx context.Context, service Service) error

	//CreateWorkdays saves new Workdays to storage.
	CreateWorkdays(ctx context.Context, workdays ...Workday) error

	// DeleteAppointmentsBeforeDate removes all appointments till the specified date for barber with specified ID.
	DeleteAppointmentsBeforeDate(ctx context.Context, barberID int64, date string) error

	// DeleteBarberByID removes barber with specified ID. It also removes all serviced, workdays and appointments associated with barber.
	DeleteBarberByID(ctx context.Context, barberID int64) error

	// DeleteServiceByID removes service with specified ID.
	DeleteServiceByID(ctx context.Context, serviceID int) error

	//DeleteWorkdaysByDateRange removes working days that fall within the date range.
	//It only removes working days for barber with specified barberID.
	DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange DateRange) error

	//FindAllBarberIDs return a slice of IDs of all barbers.
	GetAllBarberIDs(ctx context.Context) ([]int64, error)

	//GetBarberByID returns barber with barberID.
	GetBarberByID(ctx context.Context, barberID int64) (Barber, error)

	//GetLatestAppointmentDate returns the latest work date saved for barber with specified ID for which at least one appointment exists.
	//If there is no saved appointments it returns "2000-01-01".
	GetLatestAppointmentDate(ctx context.Context, barberID int64) (string, error)

	//GetLatestWorkDate returns the latest work date saved for barber with specified ID.
	//If there is no saved work dates it returns "2000-01-01".
	GetLatestWorkDate(ctx context.Context, barberID int64) (string, error)

	// GetServicesByBarberID returns all services provided by barber with specified ID.
	GetServicesByBarberID(ctx context.Context, barberID int64) ([]Service, error)

	// GetWorkdaysByDateRange returns working days that fall within the date range.
	// It only returns working days for barber with specified ID.
	// Returned working days are sorted by date in ascending order.
	GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange DateRange) ([]Workday, error)

	// Init prepares the storage for use. It creates the necessary tables if not exists.
	Init(ctx context.Context) error

	// UpdateBarber updates valid  and non-niladic fields of Barber. ID field must be non-niladic and remains not updated.
	UpdateBarber(ctx context.Context, barber Barber) error

	// UpdateService updates non-niladic fields of Service. ID field must be non-niladic and remains not updated.
	// UpdateService also doesn't updates barber_id field even if it's non-niladic.
	UpdateService(ctx context.Context, service Service) error
}

type Barber struct {
	ID   int64          `db:"id"`
	Name sql.NullString `db:"name"`

	//Format of phone number is +71234567890.
	Phone sql.NullString `db:"phone"`

	//LastWorkdate is a date in YYYY-MM-DD format in local time zone. Default is '3000-01-01'.
	LastWorkDate string `db:"last_workdate"`
}

type DateRange struct {
	StartDate string
	EndDate   string
}

type Service struct {
	//ID field autofills when new service saved. No need to pass this field to Storage: passed value will be ignored. It intended for read purposes only.
	ID         int    `db:"id"`
	BarberID   int64  `db:"barber_id"`
	Name       string `db:"name"`
	Desciption string `db:"description"`
	Price      uint   `db:"price"`
	Duration   string `db:"duration"`
}

type Workday struct {
	//ID field autofills when new workday saved. No need to pass this field to Storage: passed value will be ignored. It intended for read purposes only.
	ID       int   `db:"id"`
	BarberID int64 `db:"barber_id"`

	//Date in YYYY-MM-DD format in local time zone.
	Date string `db:"date"`

	//Beginning of the working day in HH:MM in local time zone.
	StartTime string `db:"start_time"`

	//End of the working day in HH:MM in local time zone.
	EndTime string `db:"end_time"`
}
