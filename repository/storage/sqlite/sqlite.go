package sqlite

import (
	"barbershop-bot/lib/e"
	"barbershop-bot/repository/storage"
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

// CreateBarber saves new BarberID to storage
func (s *Storage) CreateBarber(ctx context.Context, barberID int64) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := `INSERT INTO barbers (id) VALUES (?)`
	if _, err := s.db.ExecContext(ctx, q, barberID); err != nil {
		return e.Wrap("can't save barberID", err)
	}
	return nil
}

// CreateWorkdays saves new Workdays to storage
func (s *Storage) CreateWorkdays(ctx context.Context, workdays ...storage.Workday) error {
	exprs := make([]string, 0, len(workdays))
	args := make([]interface{}, 0, len(workdays))
	for _, workday := range workdays {
		exprs = append(exprs, "(?, ?, ?, ?)")
		args = append(args, workday.BarberID, workday.Date, workday.StartTime, workday.EndTime)
	}
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := fmt.Sprintf(`INSERT INTO schedule (barber_id, date, start_time, end_time) VALUES %s`, strings.Join(exprs, ", "))
	_, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		return e.Wrap("can't save workdays", err)
	}
	return nil
}

// FindAllBarberIDs return a slice of IDs of all barbers.
func (s *Storage) FindAllBarberIDs(ctx context.Context) (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	q := `SELECT id FROM barbers`
	rows, err := s.db.QueryContext(ctx, q)
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

// GetBarberByID returns status of dialog for barber with barberID.
func (s *Storage) GetBarberByID(ctx context.Context, barberID int64) (storage.Barber, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	q := `SELECT name, phone, state, state_expiration FROM barbers WHERE id = ?`
	var barber storage.Barber
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(&barber.Name, &barber.Phone, &barber.State, &barber.Expiration)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.Barber{}, storage.ErrNoSavedBarber
	}
	if err != nil {
		return storage.Barber{}, e.Wrap("can't get barber", err)
	}
	barber.ID = sql.NullInt64{Int64: barberID, Valid: true}
	return barber, nil
}

// GetLatestWorkDate returns the latest work date saved for barber with barberID.
// If there is no saved work dates it returns ("2000-01-01", ErrNoSavedWorkdates).
func (s *Storage) GetLatestWorkDate(ctx context.Context, barberID int64) (string, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	q := `SELECT COUNT(*) FROM schedule WHERE barber_id = ?`
	var count int
	if err := s.db.QueryRowContext(ctx, q, barberID).Scan(&count); err != nil {
		return "", e.Wrap("can't check if any work date exists", err)
	}
	if count == 0 {
		return "2000-01-01", storage.ErrNoSavedWorkdates
	} else {
		q = `SELECT MAX(date) FROM schedule WHERE barber_id = ?`
		var date string
		if err := s.db.QueryRowContext(ctx, q, barberID).Scan(&date); err != nil {
			return "", e.Wrap("can't get latest workdate", err)
		}
		return date, nil
	}
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
		state_expiration TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS barbers (
		id INTEGER PRIMARY KEY, 
		name TEXT UNIQUE, 
		phone TEXT UNIQUE, 
		state INTEGER,
		state_expiration TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY, 
		name TEXT UNIQUE, 
		description TEXT UNIQUE, 
		price REAL, 
		duration TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS appointments (
		id INTEGER PRIMARY KEY, 
		user_id INTEGER, 
		barber_id INTEGER, 
		service_id INTEGER, 
		date TEXT, 
		time TEXT,
		cteated_at TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS schedule (
		id INTEGER PRIMARY KEY, 
		barber_id INTEGER, 
		date TEXT, 
		start_time TEXT, 
		end_time TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

// IsBarberExists reports if barber with specified ID exists in database
func (s *Storage) IsBarberExists(ctx context.Context, barberID int64) (bool, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	q := `SELECT COUNT(*) FROM barbers WHERE id = ?`
	var count int
	if err := s.db.QueryRowContext(ctx, q, barberID).Scan(&count); err != nil {
		return false, e.Wrap("can't check if barber exists", err)
	}
	return count > 0, nil
}

// UpdateBarberName saves new name for barber with barberID.
func (s *Storage) UpdateBarberName(ctx context.Context, name string, barberID int64) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := `UPDATE barbers SET name = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, name, barberID)
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = storage.ErrNonUniqueData
		}
		return e.Wrap("can't save barber's name", err)
	}
	return nil
}

// UpdateBarberPhone saves new phone for barber with barberID.
func (s *Storage) UpdateBarberPhone(ctx context.Context, phone string, barberID int64) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := `UPDATE barbers SET phone = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, phone, barberID)
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = storage.ErrNonUniqueData
		}
		return e.Wrap("can't save barber's name", err)
	}
	return nil
}

// UpdateBarberStatus saves new status for barber with barberID.
func (s *Storage) UpdateBarberStatus(ctx context.Context, status storage.Status, barberID int64) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := `UPDATE barbers SET state = ? , state_expiration = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, status.State, status.Expiration, barberID)
	if err != nil {
		return e.Wrap("can't save barber's status", err)
	}
	return nil
}
