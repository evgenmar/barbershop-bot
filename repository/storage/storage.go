package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrAlreadyExists      = errors.New("the object being saved already exists")
	ErrAppointmentsExists = errors.New("there are active appointments for the period being deleted")
	ErrNonUniqueData      = errors.New("data to save must be unique")
	ErrNoSavedObject      = errors.New("no saved object with specified ID")
)

type Storage interface {
	// CreateAppointment saves new appointment to storage.
	// Note and ID fields are ignored. UserID field is optional.
	CreateAppointment(ctx context.Context, appt Appointment) error

	//CreateBarber saves new BarberID to storage.
	CreateBarber(ctx context.Context, barberID int64) error

	//CreateService saves new service to storage.
	CreateService(ctx context.Context, service Service) error

	// CreateUser saves new user to storage.
	CreateUser(ctx context.Context, user User) error

	//CreateWorkdays saves new Workdays to storage.
	CreateWorkdays(ctx context.Context, workdays ...Workday) error

	// DeleteAppointmentByID removes appointment with specified ID.
	DeleteAppointmentByID(ctx context.Context, appointmentID int) error

	// DeleteBarberByID removes barber with specified ID. It also removes all serviced, workdays and appointments associated with barber.
	DeleteBarberByID(ctx context.Context, barberID int64) error

	// DeletePastAppointments removes appointments at dates before today for barber with specified ID.
	DeletePastAppointments(ctx context.Context, barberID int64) error

	// DeleteServiceByID removes service with specified ID.
	DeleteServiceByID(ctx context.Context, serviceID int) error

	//DeleteWorkdaysByDateRange removes working days that fall within the date range.
	//It only removes working days for barber with specified barberID.
	DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange DateRange) error

	// GetAllBarbers return a slice of all barbers.
	GetAllBarbers(ctx context.Context) ([]Barber, error)

	// GetAppointmentByID returns appointment with specified ID.
	GetAppointmentByID(ctx context.Context, appointmentID int) (Appointment, error)

	// GetAppointmentIDByWorkdayIDAndTime returns ID of appointment with specified workdayID and time.
	GetAppointmentIDByWorkdayIDAndTime(ctx context.Context, workdayID int, time int16) (int, error)

	// GetAppointmentsByDateRange returns appointments that fall within the date range.
	// It only returns appointments for barber with specified ID.
	// Returned appointments are sorted by date and time in ascending order.
	GetAppointmentsByDateRange(ctx context.Context, barberID int64, dateRange DateRange) ([]Appointment, error)

	//GetBarberByID returns barber with specified ID.
	GetBarberByID(ctx context.Context, barberID int64) (Barber, error)

	//GetLatestAppointmentDate returns the latest work date saved for barber with specified ID for which at least one appointment exists.
	//If there is no saved appointments it returns "2000-01-01".
	GetLatestAppointmentDate(ctx context.Context, barberID int64) (string, error)

	//GetLatestWorkDate returns the latest work date saved for barber with specified ID.
	//If there is no saved work dates it returns "2000-01-01".
	GetLatestWorkDate(ctx context.Context, barberID int64) (string, error)

	// GetServiceByID returns service with specified ID.
	GetServiceByID(ctx context.Context, serviceID int) (Service, error)

	// GetServicesByBarberID returns all services provided by barber with specified ID.
	// Returned services are sorted by price in ascending order.
	GetServicesByBarberID(ctx context.Context, barberID int64) ([]Service, error)

	// GetUpcomingAppointment returns an upcoming appointment for user with specified ID.
	GetUpcomingAppointment(ctx context.Context, userID int64) (Appointment, error)

	// GetUserByID returns user with specified ID.
	GetUserByID(ctx context.Context, userID int64) (User, error)

	// GetWorkdayByID returns workday with specified ID.
	GetWorkdayByID(ctx context.Context, workdayID int) (Workday, error)

	// GetWorkdaysByDateRange returns working days that fall within the date range.
	// It only returns working days for barber with specified ID.
	// Returned working days are sorted by date in ascending order.
	GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange DateRange) ([]Workday, error)

	// Init prepares the storage for use. It creates the necessary tables if not exists.
	Init(ctx context.Context) error

	// UpdateAppointment updates non-niladic fields of Appointment. ID field must be non-niladic and remains not updated.
	// UpdateAppointment also doesn't updates UserID, ServiceID, Duration, CreatedAt fields even if non-niladic.
	UpdateAppointment(ctx context.Context, appointment Appointment) error

	// UpdateBarber updates valid and non-niladic fields of Barber. ID field must be non-niladic and remains not updated.
	UpdateBarber(ctx context.Context, barber Barber) error

	// UpdateService updates non-niladic fields of Service. ID field must be non-niladic and remains not updated.
	// UpdateService also doesn't updates BarberID field even if it's non-niladic.
	UpdateService(ctx context.Context, service Service) error

	// UpdateUser updates valid fields of User. ID field must be non-niladic and remains not updated.
	UpdateUser(ctx context.Context, user User) error
}

type Appointment struct {
	ID        int            `db:"id"`
	UserID    sql.NullInt64  `db:"user_id"`
	WorkdayID int            `db:"workday_id"`
	ServiceID sql.NullInt32  `db:"service_id"`
	Time      int16          `db:"time"`
	Duration  int16          `db:"duration"`
	Note      sql.NullString `db:"note"`

	//CreatedAt has a format of Unix time
	CreatedAt int64 `db:"created_at"`
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
	Duration   int16  `db:"duration"`
}

type User struct {
	ID   int64          `db:"id"`
	Name sql.NullString `db:"name"`

	//Format of phone number is +71234567890.
	Phone sql.NullString `db:"phone"`
}

type Workday struct {
	//ID field autofills when new workday saved. No need to pass this field to Storage: passed value will be ignored. It intended for read purposes only.
	ID       int   `db:"id"`
	BarberID int64 `db:"barber_id"`

	//Date in YYYY-MM-DD format in local time zone.
	Date string `db:"date"`

	//Beginning of the working day in HH:MM in local time zone.
	StartTime int16 `db:"start_time"`

	//End of the working day in HH:MM in local time zone.
	EndTime int16 `db:"end_time"`
}
