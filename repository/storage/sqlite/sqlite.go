package sqlite

import (
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	st "barbershop-bot/repository/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

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

// CreateBarber saves new BarberID to storage.
func (s *Storage) CreateBarber(ctx context.Context, barberID int64) error {
	q := `INSERT INTO barbers (id) VALUES (?)`
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, barberID)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrAlreadyExists
		}
		return e.Wrap("can't save barberID", err)
	}
	return nil
}

// CreateService saves new service to storage.
func (s *Storage) CreateService(ctx context.Context, service st.Service) (err error) {
	defer func() { err = e.WrapIfErr("can't save service", err) }()
	if service.BarberID < 1 || service.Name == "" || service.Desciption == "" || service.Price < 1 || service.Duration == "" {
		return st.ErrInvalidService
	}
	q := `INSERT INTO services (barber_id, name, description, price, duration) VALUES (?, ?, ?, ?, ?)`
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, service.BarberID, service.Name, service.Desciption, service.Price, service.Duration)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// CreateUser saves new user to storage.
func (s *Storage) CreateUser(ctx context.Context, user st.User) (err error) {
	defer func() { err = e.WrapIfErr("can't save user", err) }()
	columns := ""
	placeholders := ""
	args := make([]interface{}, 0, 3)
	args = append(args, user.ID)
	if user.Name.Valid {
		columns += ", name"
		placeholders += ", ?"
		args = append(args, user.Name)
	}
	if user.Phone.Valid {
		columns += ", phone"
		placeholders += ", ?"
		args = append(args, user.Phone)
	}
	q := fmt.Sprintf(`INSERT INTO users (id%s) VALUES (?%s)`, columns, placeholders)
	s.rwMutex.Lock()
	_, err = s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// CreateWorkdays saves new Workdays to storage.
func (s *Storage) CreateWorkdays(ctx context.Context, workdays ...st.Workday) error {
	placeholders := make([]string, 0, len(workdays))
	args := make([]interface{}, 0, len(workdays))
	for _, workday := range workdays {
		placeholders = append(placeholders, "(?, ?, ?, ?)")
		args = append(args, workday.BarberID, workday.Date, workday.StartTime, workday.EndTime)
	}
	q := fmt.Sprintf(`INSERT INTO workdays (barber_id, date, start_time, end_time) VALUES %s`, strings.Join(placeholders, ", "))
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrAlreadyExists
		}
		return e.Wrap("can't save workdays", err)
	}
	return nil
}

// DeleteAppointmentsBeforeDate removes all appointments till the specified date for barber with specified ID.
func (s *Storage) DeleteAppointmentsBeforeDate(ctx context.Context, barberID int64, date string) error {
	q := `DELETE FROM appointments WHERE workday_id IN (
		SELECT id FROM workdays WHERE barber_id = ? AND date < ?
		)`
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, barberID, date)
	s.rwMutex.Unlock()
	if err != nil {
		return e.Wrap("can't delete appointments", err)
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

// DeleteWorkdaysByDateRange removes working days that fall within the date range.
// It only removes working days for barber with specified barberID.
func (s *Storage) DeleteWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange st.DateRange) error {
	q := `DELETE FROM workdays WHERE barber_id = ? AND date BETWEEN ? AND ?`
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, barberID, dateRange.StartDate, dateRange.EndDate)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrAppointmentsExists
		}
		return e.Wrap("can't delete workdays", err)
	}
	return nil
}

// GetAllBarbers return a slice of all barbers.
func (s *Storage) GetAllBarbers(ctx context.Context) (barbers []st.Barber, err error) {
	defer func() { err = e.WrapIfErr("can't get barbers", err) }()
	q := `SELECT id, name, phone, last_workdate FROM barbers`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q)
	s.rwMutex.RUnlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var barber st.Barber
		if err := rows.Scan(&barber.ID, &barber.Name, &barber.Phone, &barber.LastWorkDate); err != nil {
			return nil, err
		}
		barbers = append(barbers, barber)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return barbers, nil
}

// GetBarberByID returns barber with specified ID.
func (s *Storage) GetBarberByID(ctx context.Context, barberID int64) (st.Barber, error) {
	q := `SELECT name, phone, last_workdate FROM barbers WHERE id = ?`
	var barber st.Barber
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(&barber.Name, &barber.Phone, &barber.LastWorkDate)
	s.rwMutex.RUnlock()
	if errors.Is(err, sql.ErrNoRows) {
		return st.Barber{}, st.ErrNoSavedBarber
	}
	if err != nil {
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
func (s *Storage) GetServiceByID(ctx context.Context, serviceID int) (st.Service, error) {
	q := `SELECT barber_id, name, description, price, duration FROM services WHERE id = ?`
	var service st.Service
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, serviceID).Scan(&service.BarberID, &service.Name, &service.Desciption, &service.Price, &service.Duration)
	s.rwMutex.RUnlock()
	if errors.Is(err, sql.ErrNoRows) {
		return st.Service{}, st.ErrNoSavedService
	}
	if err != nil {
		return st.Service{}, e.Wrap("can't get service", err)
	}
	service.ID = serviceID
	return service, nil
}

// GetServicesByBarberID returns all services provided by barber with specified ID.
func (s *Storage) GetServicesByBarberID(ctx context.Context, barberID int64) (services []st.Service, err error) {
	defer func() { err = e.WrapIfErr("can't get services", err) }()
	q := `SELECT id, name, description, price, duration FROM services WHERE barber_id = ?`
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

// GetUserByID returns user with specified ID.
func (s *Storage) GetUserByID(ctx context.Context, userID int64) (st.User, error) {
	q := `SELECT name, phone FROM users WHERE id = ?`
	var user st.User
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&user.Name, &user.Phone)
	s.rwMutex.RUnlock()
	if errors.Is(err, sql.ErrNoRows) {
		return st.User{}, st.ErrNoSavedUser
	}
	if err != nil {
		return st.User{}, e.Wrap("can't get user", err)
	}
	user.ID = userID
	return user, nil
}

// GetWorkdaysByDateRange returns working days that fall within the date range.
// It only returns working days for barber with specified ID.
// Returned working days are sorted by date in ascending order.
func (s *Storage) GetWorkdaysByDateRange(ctx context.Context, barberID int64, dateRange st.DateRange) (workdays []st.Workday, err error) {
	defer func() { err = e.WrapIfErr("can't get workdays", err) }()
	q := `SELECT id, date, start_time, end_time FROM workdays WHERE barber_id = ? AND date BETWEEN ? AND ? ORDER BY date ASC`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q, barberID, dateRange.StartDate, dateRange.EndDate)
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
		phone TEXT
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS barbers (
		id INTEGER PRIMARY KEY, 
		name TEXT UNIQUE, 
		phone TEXT UNIQUE, 
		last_workdate TEXT DEFAULT '` + cfg.InfiniteWorkDate + `'
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
		duration TEXT NOT NULL,
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
		start_time TEXT NOT NULL, 
		end_time TEXT NOT NULL,
		UNIQUE (barber_id, date),
		FOREIGN KEY (barber_id) REFERENCES barbers(id) ON DELETE CASCADE
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS appointments (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		workday_id INTEGER NOT NULL,
		service_id INTEGER,
		time TEXT NOT NULL,
		duration TEXT NOT NULL,
		cteated_at TEXT NOT NULL,
		UNIQUE (workday_id, time),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (workday_id) REFERENCES workdays(id) ON DELETE RESTRICT,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

// UpdateBarber updates valid and non-niladic fields of Barber. ID field must be non-niladic and remains not updated.
func (s *Storage) UpdateBarber(ctx context.Context, barber st.Barber) error {
	query := make([]string, 0, 3)
	args := make([]interface{}, 0, 3)
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
	args = append(args, barber.ID)
	q := fmt.Sprintf(`UPDATE barbers SET %s WHERE id = ?`, strings.Join(query, " , "))
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrNonUniqueData
		}
		return e.Wrap("can't update barber", err)
	}
	return nil
}

// UpdateService updates non-niladic fields of Service. ID field must be non-niladic and remains not updated.
// UpdateService also doesn't updates barber_id field even if it's non-niladic.
func (s *Storage) UpdateService(ctx context.Context, service st.Service) (err error) {
	defer func() { err = e.WrapIfErr("can't update service", err) }()
	if service.ID < 1 {
		return st.ErrUnspecifiedServiceID
	}
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
	if service.Duration != "" {
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
			err = st.ErrNonUniqueData
		}
		return err
	}
	return nil
}

// UpdateUser updates valid fields of Barber. ID field must be non-niladic and remains not updated.
func (s *Storage) UpdateUser(ctx context.Context, user st.User) error {
	query := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if user.Name.Valid {
		query = append(query, "name = ?")
		args = append(args, user.Name)
	}
	if user.Phone.Valid {
		query = append(query, "phone = ?")
		args = append(args, user.Phone)
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
