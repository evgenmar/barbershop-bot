package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	cfg "github.com/evgenmar/barbershop-bot/lib/config"
	"github.com/evgenmar/barbershop-bot/lib/e"
	tm "github.com/evgenmar/barbershop-bot/lib/time"
	m "github.com/evgenmar/barbershop-bot/repository/mappers"
	st "github.com/evgenmar/barbershop-bot/repository/storage"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Storage struct {
	db      *sql.DB
	rwMutex sync.RWMutex
}

// New creates a new SQLite storage.
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, e.Wrap("can't open database", err)
	}
	if err := db.Ping(); err != nil {
		return nil, e.Wrap("can't connect to database", err)
	}
	return &Storage{db: db}, nil
}

// CreateAppointment saves new appointment to storage.
// Note and ID fields are ignored. UserID field is optional.
func (s *Storage) CreateAppointment(ctx context.Context, appt st.Appointment) (err error) {
	defer func() { err = e.WrapIfErr("can't save appointment", err) }()
	var q string
	args := make([]interface{}, 0, 6)
	if appt.UserID.Valid {
		q = `INSERT INTO appointments (user_id, workday_id, service_id, time, duration, created_at) VALUES (?, ?, ?, ?, ?, ?)`
		args = append(args, appt.UserID, appt.WorkdayID, appt.ServiceID, appt.Time, appt.Duration, appt.CreatedAt)
	} else {
		q = `INSERT INTO appointments (workday_id, service_id, time, duration, created_at) VALUES (?, ?, ?, ?, ?)`
		args = append(args, appt.WorkdayID, appt.ServiceID, appt.Time, appt.Duration, appt.CreatedAt)
	}
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// CreateBarber saves new BarberID to storage.
func (s *Storage) CreateBarber(ctx context.Context, barberID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't save barberID", err) }()
	q := `INSERT INTO barbers (id) VALUES (?)`
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, barberID)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// CreateService saves new service to storage.
func (s *Storage) CreateService(ctx context.Context, service st.Service) (err error) {
	defer func() { err = e.WrapIfErr("can't save service", err) }()
	q := `INSERT INTO services (barber_id, name, description, price, duration) VALUES (?, ?, ?, ?, ?)`
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, service.BarberID, service.Name, service.Desciption, service.Price, service.Duration)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// CreateUser saves new user to storage.
func (s *Storage) CreateUser(ctx context.Context, user st.User) (err error) {
	defer func() { err = e.WrapIfErr("can't save user", err) }()
	q := `INSERT INTO users (id, message_id, chat_id) VALUES (?, ?, ?)`
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, user.ID, user.MessageID, user.ChatID)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// CreateWorkdays saves new Workdays to storage.
func (s *Storage) CreateWorkdays(ctx context.Context, workdays ...st.Workday) (err error) {
	defer func() { err = e.WrapIfErr("can't save workdays", err) }()
	placeholders := make([]string, 0, len(workdays))
	args := make([]interface{}, 0, len(workdays))
	for _, workday := range workdays {
		placeholders = append(placeholders, "(?, ?, ?, ?)")
		args = append(args, workday.BarberID, workday.Date, workday.StartTime, workday.EndTime)
	}
	q := fmt.Sprintf(`INSERT INTO workdays (barber_id, date, start_time, end_time) VALUES %s`, strings.Join(placeholders, ", "))
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// DeleteAppointmentByID removes appointment with specified ID.
func (s *Storage) DeleteAppointmentByID(ctx context.Context, appointmentID int) error {
	q := `DELETE FROM appointments WHERE id = ?`
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, appointmentID)
	s.rwMutex.Unlock()
	if err != nil {
		return e.Wrap("can't delete appointment", err)
	}
	return nil
}

// DeleteBarberByID removes barber with specified ID. It also removes all serviced, workdays and appointments associated with barber.
func (s *Storage) DeleteBarberByID(ctx context.Context, barberID int64) error {
	q := `DELETE FROM barbers WHERE id = ?`
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, barberID)
	s.rwMutex.Unlock()
	if err != nil {
		return e.Wrap("can't delete barber", err)
	}
	return nil
}

// DeletePastAppointments removes appointments at dates before today for barber with specified ID.
func (s *Storage) DeletePastAppointments(ctx context.Context, barberID int64) error {
	q := `DELETE FROM appointments WHERE workday_id IN
	(SELECT id FROM workdays WHERE barber_id = ? AND date < ?)`
	today := m.MapToStorage.Date(tm.Today())
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, barberID, today)
	s.rwMutex.Unlock()
	if err != nil {
		return e.Wrap("can't delete appointments", err)
	}
	return nil
}

// DeleteServiceByID removes service with specified ID.
func (s *Storage) DeleteServiceByID(ctx context.Context, serviceID int) error {
	q := `DELETE FROM services WHERE id = ?`
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, serviceID)
	s.rwMutex.Unlock()
	if err != nil {
		return e.Wrap("can't delete service", err)
	}
	return nil
}

// DeleteWorkdayByID removes workday with specified ID.
func (s *Storage) DeleteWorkdayByID(ctx context.Context, workdayID int) (err error) {
	defer func() { err = e.WrapIfErr("can't delete workday", err) }()
	q := `DELETE FROM workdays WHERE id = ?`
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, workdayID)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAppointmentsExists
		}
		return err
	}
	return nil
}

// DeleteWorkdaysByDateRange removes working days that fall within the date range.
// It only removes working days for barber with specified barberID.
func (s *Storage) DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange st.DateRange) (err error) {
	defer func() { err = e.WrapIfErr("can't delete workdays", err) }()
	q := `DELETE FROM workdays WHERE barber_id = ? AND date BETWEEN ? AND ?`
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, barberID, dateRange.FirstDate, dateRange.LastDate)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrAppointmentsExists
		}
		return err
	}
	return nil
}

// GetAllBarbers return a slice of all barbers.
func (s *Storage) GetAllBarbers(ctx context.Context) (barbers []st.Barber, err error) {
	defer func() { err = e.WrapIfErr("can't get barbers", err) }()
	q := `SELECT id, name, phone, last_workdate, message_id, chat_id FROM barbers`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q)
	s.rwMutex.RUnlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var barber st.Barber
		if err := rows.Scan(&barber.ID, &barber.Name, &barber.Phone, &barber.LastWorkDate, &barber.MessageID, &barber.ChatID); err != nil {
			return nil, err
		}
		barbers = append(barbers, barber)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return barbers, nil
}

// GetAppointmentByID returns appointment with specified ID.
func (s *Storage) GetAppointmentByID(ctx context.Context, appointmentID int) (appointment st.Appointment, err error) {
	defer func() { err = e.WrapIfErr("can't get appointment", err) }()
	q := `SELECT user_id, workday_id, service_id, time, duration, note, created_at FROM appointments WHERE id = ?`
	s.rwMutex.RLock()
	err = s.db.QueryRowContext(ctx, q, appointmentID).Scan(
		&appointment.UserID,
		&appointment.WorkdayID,
		&appointment.ServiceID,
		&appointment.Time,
		&appointment.Duration,
		&appointment.Note,
		&appointment.CreatedAt,
	)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return st.Appointment{}, st.ErrNoSavedObject
		}
		return st.Appointment{}, err
	}
	appointment.ID = appointmentID
	return appointment, nil
}

// GetAppointmentIDByWorkdayIDAndTime returns ID of appointment with specified workdayID and time.
func (s *Storage) GetAppointmentIDByWorkdayIDAndTime(ctx context.Context, workdayID int, time int16) (appointmentID int, err error) {
	defer func() { err = e.WrapIfErr("can't get appointmentID", err) }()
	q := `SELECT id FROM appointments WHERE workday_id = ? AND time = ?`
	s.rwMutex.RLock()
	err = s.db.QueryRowContext(ctx, q, workdayID, time).Scan(&appointmentID)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, st.ErrNoSavedObject
		}
		return 0, err
	}
	return appointmentID, nil
}

// GetAppointmentsByDateRange returns appointments that fall within the date range.
// It only returns appointments for barber with specified ID.
// Returned appointments are sorted by date and time in ascending order.
func (s *Storage) GetAppointmentsByDateRange(ctx context.Context, barberID int64, dateRange st.DateRange) (appointments []st.Appointment, err error) {
	defer func() { err = e.WrapIfErr("can't get appointments", err) }()
	q := `SELECT a.id, a.user_id, a.workday_id, a.service_id, a.time, a.duration, a.note, a.created_at, w.date 
		FROM appointments a JOIN workdays w ON a.workday_id = w.id
		WHERE w.barber_id = ? AND w.date BETWEEN ? AND ? ORDER BY w.date, a.time`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q, barberID, dateRange.FirstDate, dateRange.LastDate)
	s.rwMutex.RUnlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ignored sql.RawBytes
	for rows.Next() {
		var a st.Appointment
		if err := rows.Scan(&a.ID, &a.UserID, &a.WorkdayID, &a.ServiceID, &a.Time, &a.Duration, &a.Note, &a.CreatedAt, &ignored); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return appointments, nil
}

// GetBarberByID returns barber with specified ID.
func (s *Storage) GetBarberByID(ctx context.Context, barberID int64) (st.Barber, error) {
	q := `SELECT name, phone, last_workdate, message_id, chat_id FROM barbers WHERE id = ?`
	var barber st.Barber
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(
		&barber.Name,
		&barber.Phone,
		&barber.LastWorkDate,
		&barber.MessageID,
		&barber.ChatID,
	)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return st.Barber{}, st.ErrNoSavedObject
		}
		return st.Barber{}, e.Wrap("can't get barber", err)
	}
	barber.ID = barberID
	return barber, nil
}

// GetLatestAppointmentDate returns the latest work date saved for barber with specified ID for which at least one appointment exists.
// If there is no saved appointments it returns "2000-01-01".
func (s *Storage) GetLatestAppointmentDate(ctx context.Context, barberID int64) (string, error) {
	q := `SELECT workdays.date FROM workdays INNER JOIN appointments ON workdays.id = appointments.workday_id
WHERE workdays.barber_id = ? ORDER BY workdays.date DESC LIMIT 1`
	var date string
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(&date)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "2000-01-01", nil
		}
		return "", e.Wrap("can't get latest appointment date", err)
	}
	return date, nil
}

// GetLatestWorkDate returns the latest work date saved for barber with specified ID.
// If there is no saved work dates it returns "2000-01-01".
func (s *Storage) GetLatestWorkDate(ctx context.Context, barberID int64) (string, error) {
	q := `SELECT date FROM workdays WHERE barber_id = ? ORDER BY date DESC LIMIT 1`
	var date string
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(&date)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "2000-01-01", nil
		}
		return "", e.Wrap("can't get latest workdate", err)
	}
	return date, nil
}

// GetServiceByID returns service with specified ID.
func (s *Storage) GetServiceByID(ctx context.Context, serviceID int) (service st.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get service", err) }()
	q := `SELECT barber_id, name, description, price, duration FROM services WHERE id = ?`
	s.rwMutex.RLock()
	err = s.db.QueryRowContext(ctx, q, serviceID).Scan(
		&service.BarberID,
		&service.Name,
		&service.Desciption,
		&service.Price,
		&service.Duration,
	)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return st.Service{}, st.ErrNoSavedObject
		}
		return st.Service{}, err
	}
	service.ID = serviceID
	return service, nil
}

// GetServicesByBarberID returns all services provided by barber with specified ID.
// Returned services are sorted by price in ascending order.
func (s *Storage) GetServicesByBarberID(ctx context.Context, barberID int64) (services []st.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get services", err) }()
	q := `SELECT id, name, description, price, duration FROM services WHERE barber_id = ? ORDER BY price`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q, barberID)
	s.rwMutex.RUnlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var service st.Service
		if err := rows.Scan(&service.ID, &service.Name, &service.Desciption, &service.Price, &service.Duration); err != nil {
			return nil, err
		}
		service.BarberID = barberID
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return services, nil
}

// GetUpcomingAppointment returns an upcoming appointment for user with specified ID.
func (s *Storage) GetUpcomingAppointment(ctx context.Context, userID int64) (appointment st.Appointment, err error) {
	defer func() { err = e.WrapIfErr("can't get appointment", err) }()
	q := `SELECT a.id, a.workday_id, a.service_id, a.time, a.duration, a.note, a.created_at, w.date 
		FROM appointments a JOIN workdays w ON a.workday_id = w.id
		WHERE a.user_id = ? AND (w.date > ? OR (w.date = ? AND a.time + a.duration > ?))`
	today := m.MapToStorage.Date(tm.Today())
	currentTime, err := m.MapToStorage.Duration(tm.CurrentDayTime())
	if err != nil {
		return st.Appointment{}, err
	}
	var date string
	s.rwMutex.RLock()
	err = s.db.QueryRowContext(ctx, q, userID, today, today, currentTime).Scan(
		&appointment.ID,
		&appointment.WorkdayID,
		&appointment.ServiceID,
		&appointment.Time,
		&appointment.Duration,
		&appointment.Note,
		&appointment.CreatedAt,
		&date,
	)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return st.Appointment{}, st.ErrNoSavedObject
		}
		return st.Appointment{}, err
	}
	appointment.UserID = sql.NullInt64{Int64: userID, Valid: true}
	return appointment, nil
}

// GetUserByID returns user with specified ID.
func (s *Storage) GetUserByID(ctx context.Context, userID int64) (user st.User, err error) {
	defer func() { err = e.WrapIfErr("can't get user", err) }()
	q := `SELECT name, phone, message_id, chat_id FROM users WHERE id = ?`
	s.rwMutex.RLock()
	err = s.db.QueryRowContext(ctx, q, userID).Scan(&user.Name, &user.Phone, &user.MessageID, &user.ChatID)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return st.User{}, st.ErrNoSavedObject
		}
		return st.User{}, err
	}
	user.ID = userID
	return user, nil
}

// GetWorkdayByID returns workday with specified ID.
func (s *Storage) GetWorkdayByID(ctx context.Context, workdayID int) (workday st.Workday, err error) {
	defer func() { err = e.WrapIfErr("can't get workday", err) }()
	q := `SELECT barber_id, date, start_time, end_time FROM workdays WHERE id = ?`
	s.rwMutex.RLock()
	err = s.db.QueryRowContext(ctx, q, workdayID).Scan(&workday.BarberID, &workday.Date, &workday.StartTime, &workday.EndTime)
	s.rwMutex.RUnlock()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return st.Workday{}, st.ErrNoSavedObject
		}
		return st.Workday{}, err
	}
	workday.ID = workdayID
	return workday, nil
}

// GetWorkdaysByDateRange returns working days that fall within the date range.
// It only returns working days for barber with specified ID.
// Returned working days are sorted by date in ascending order.
func (s *Storage) GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange st.DateRange) (workdays []st.Workday, err error) {
	defer func() { err = e.WrapIfErr("can't get workdays", err) }()
	q := `SELECT id, date, start_time, end_time FROM workdays WHERE barber_id = ? AND date BETWEEN ? AND ? ORDER BY date ASC`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q, barberID, dateRange.FirstDate, dateRange.LastDate)
	s.rwMutex.RUnlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var workday st.Workday
		if err := rows.Scan(&workday.ID, &workday.Date, &workday.StartTime, &workday.EndTime); err != nil {
			return nil, err
		}
		workday.BarberID = barberID
		workdays = append(workdays, workday)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return workdays, nil
}

// Init prepares the storage for use. It creates the necessary tables if not exists.
func (s *Storage) Init(ctx context.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't create table", err) }()
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY, 
		name TEXT, 
		phone TEXT,
		message_id TEXT NOT NULL,
		chat_id INTEGER NOT NULL
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS barbers (
		id INTEGER PRIMARY KEY, 
		name TEXT UNIQUE, 
		phone TEXT UNIQUE, 
		last_workdate TEXT DEFAULT '` + cfg.InfiniteWorkDate + `',
		message_id TEXT DEFAULT '',
		chat_id INTEGER DEFAULT 0
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY, 
		barber_id INTEGER NOT NULL,
		name TEXT NOT NULL, 
		description TEXT NOT NULL, 
		price INTEGER NOT NULL, 
		duration INTEGER NOT NULL,
		UNIQUE (barber_id, name),
		FOREIGN KEY (barber_id) REFERENCES barbers(id) ON DELETE CASCADE
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS workdays (
		id INTEGER PRIMARY KEY, 
		barber_id INTEGER NOT NULL, 
		date TEXT NOT NULL, 
		start_time INTEGER NOT NULL, 
		end_time INTEGER NOT NULL,
		UNIQUE (barber_id, date),
		FOREIGN KEY (barber_id) REFERENCES barbers(id) ON DELETE CASCADE
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS appointments (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		workday_id INTEGER NOT NULL,
		service_id INTEGER,
		time INTEGER NOT NULL,
		duration INTEGER NOT NULL,
		note TEXT,
		created_at INTEGER NOT NULL,
		UNIQUE (workday_id, time),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (workday_id) REFERENCES workdays(id) ON DELETE RESTRICT,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE SET NULL
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

// UpdateAppointment updates non-niladic fields of Appointment. ID field must be non-niladic and remains not updated.
// UpdateAppointment also doesn't updates UserID, ServiceID, Duration, CreatedAt fields even if non-niladic.
func (s *Storage) UpdateAppointment(ctx context.Context, appointment st.Appointment) (err error) {
	defer func() { err = e.WrapIfErr("can't update appointment", err) }()
	query := make([]string, 0, 3)
	args := make([]interface{}, 0, 3)
	if appointment.WorkdayID != 0 {
		query = append(query, "workday_id = ?")
		args = append(args, appointment.WorkdayID)
	}
	if appointment.Time != 0 {
		query = append(query, "time = ?")
		args = append(args, appointment.Time)
	}
	if appointment.Note.Valid {
		query = append(query, "note = ?")
		args = append(args, appointment.Note)
	}
	args = append(args, appointment.ID)
	q := fmt.Sprintf(`UPDATE appointments SET %s WHERE id = ?`, strings.Join(query, " , "))
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrNonUniqueData
		}
		return err
	}
	return nil
}

// UpdateBarber updates valid and non-niladic fields of Barber. ID field must be non-niladic and remains not updated.
func (s *Storage) UpdateBarber(ctx context.Context, barber st.Barber) (err error) {
	defer func() { err = e.WrapIfErr("can't update barber", err) }()
	query := make([]string, 0, 5)
	args := make([]interface{}, 0, 5)
	if barber.Name.Valid {
		query = append(query, "name = ?")
		args = append(args, barber.Name)
	}
	if barber.Phone.Valid {
		query = append(query, "phone = ?")
		args = append(args, barber.Phone)
	}
	if barber.LastWorkDate != "" {
		query = append(query, "last_workdate = ?")
		args = append(args, barber.LastWorkDate)
	}
	if barber.MessageID != "" {
		query = append(query, "message_id = ?")
		args = append(args, barber.MessageID)
	}
	if barber.ChatID != 0 {
		query = append(query, "chat_id = ?")
		args = append(args, barber.ChatID)
	}
	args = append(args, barber.ID)
	q := fmt.Sprintf(`UPDATE barbers SET %s WHERE id = ?`, strings.Join(query, " , "))
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrNonUniqueData
		}
		return err
	}
	return nil
}

// UpdateService updates non-niladic fields of Service. ID field must be non-niladic and remains not updated.
// UpdateService also doesn't updates BarberID field even if it's non-niladic.
func (s *Storage) UpdateService(ctx context.Context, service st.Service) (err error) {
	defer func() { err = e.WrapIfErr("can't update service", err) }()
	query := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)
	if service.Name != "" {
		query = append(query, "name = ?")
		args = append(args, service.Name)
	}
	if service.Desciption != "" {
		query = append(query, "description = ?")
		args = append(args, service.Desciption)
	}
	if service.Price != 0 {
		query = append(query, "price = ?")
		args = append(args, service.Price)
	}
	if service.Duration != 0 {
		query = append(query, "duration = ?")
		args = append(args, service.Duration)
	}
	args = append(args, service.ID)
	q := fmt.Sprintf(`UPDATE services SET %s WHERE id = ?`, strings.Join(query, " , "))
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			return st.ErrNonUniqueData
		}
		return err
	}
	return nil
}

// UpdateUser updates valid fields of User. ID field must be non-niladic and remains not updated.
func (s *Storage) UpdateUser(ctx context.Context, user st.User) error {
	query := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)
	if user.Name.Valid {
		query = append(query, "name = ?")
		args = append(args, user.Name)
	}
	if user.Phone.Valid {
		query = append(query, "phone = ?")
		args = append(args, user.Phone)
	}
	if user.MessageID != "" {
		query = append(query, "message_id = ?")
		args = append(args, user.MessageID)
	}
	if user.ChatID != 0 {
		query = append(query, "chat_id = ?")
		args = append(args, user.ChatID)
	}
	args = append(args, user.ID)
	q := fmt.Sprintf(`UPDATE users SET %s WHERE id = ?`, strings.Join(query, " , "))
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		return e.Wrap("can't update user", err)
	}
	return nil
}

// UpdateWorkday updates non-niladic fields of Workday. ID field must be non-niladic and remains not updated.
// UpdateWorkday also doesn't updates BarberID and Date fields even if non-niladic.
func (s *Storage) UpdateWorkday(ctx context.Context, workday st.Workday) (err error) {
	defer func() { err = e.WrapIfErr("can't update workday", err) }()
	query := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if workday.StartTime != 0 {
		query = append(query, "start_time = ?")
		args = append(args, workday.StartTime)
	}
	if workday.EndTime != 0 {
		query = append(query, "end_time = ?")
		args = append(args, workday.EndTime)
	}
	args = append(args, workday.ID)
	q := fmt.Sprintf(`UPDATE workdays SET %s WHERE id = ?`, strings.Join(query, " , "))
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		return err
	}
	return nil
}
