package sqlite

import (
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

// FindAllBarberIDs return a slice of IDs of all barbers.
func (s *Storage) FindAllBarberIDs(ctx context.Context) (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()
	q := `SELECT id FROM barbers`
	s.rwMutex.RLock()
	rows, err := s.db.QueryContext(ctx, q)
	s.rwMutex.RUnlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var barberID int64
		if err := rows.Scan(&barberID); err != nil {
			return nil, err
		}
		barberIDs = append(barberIDs, barberID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return barberIDs, nil
}

// GetBarberByID returns barber with barberID.
func (s *Storage) GetBarberByID(ctx context.Context, barberID int64) (st.Barber, error) {
	q := `SELECT name, phone, last_workdate, state, state_expiration FROM barbers WHERE id = ?`
	var barber st.Barber
	s.rwMutex.RLock()
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(&barber.Name, &barber.Phone, &barber.LastWorkDate, &barber.State, &barber.Expiration)
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
		phone TEXT, 
		state INTEGER,
		state_expiration TEXT
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS barbers (
		id INTEGER PRIMARY KEY, 
		name TEXT UNIQUE, 
		phone TEXT UNIQUE, 
		last_workdate TEXT DEFAULT '3000-01-01',
		state INTEGER,
		state_expiration TEXT
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

// UpdateBarber updates valid fields of Barber. ID field must be valid.
func (s *Storage) UpdateBarber(ctx context.Context, barber st.Barber) error {
	query := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)
	if barber.Name.Valid {
		query = append(query, "name = ? ")
		args = append(args, barber.Name)
	}
	if barber.Phone.Valid {
		query = append(query, "phone = ? ")
		args = append(args, barber.Phone)
	}
	if barber.LastWorkDate != "" {
		query = append(query, "last_workdate = ? ")
		args = append(args, barber.LastWorkDate)
	}
	if barber.State.Valid && barber.Expiration.Valid {
		query = append(query, "state = ? , state_expiration = ?")
		args = append(args, barber.State, barber.Expiration)
	}
	args = append(args, barber.ID)
	q := fmt.Sprintf(`UPDATE barbers SET %s WHERE id = ?`, strings.Join(query, ", "))
	s.rwMutex.Lock()
	_, err := s.db.ExecContext(ctx, q, args...)
	s.rwMutex.Unlock()
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrNonUniqueData
		}
		return e.Wrap("can't save barber", err)
	}
	return nil
}
