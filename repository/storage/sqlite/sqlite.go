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
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := `INSERT INTO barbers (id) VALUES (?)`
	if _, err := s.db.ExecContext(ctx, q, barberID); err != nil {
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
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	q := fmt.Sprintf(`INSERT INTO schedule (barber_id, date, start_time, end_time) VALUES %s`, strings.Join(placeholders, ", "))
	_, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrAlreadyExists
		}
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
func (s *Storage) GetBarberByID(ctx context.Context, barberID int64) (st.Barber, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	q := `SELECT name, phone, state, state_expiration FROM barbers WHERE id = ?`
	var barber st.Barber
	err := s.db.QueryRowContext(ctx, q, barberID).Scan(&barber.Name, &barber.Phone, &barber.State, &barber.Expiration)
	if errors.Is(err, sql.ErrNoRows) {
		return st.Barber{}, st.ErrNoSavedBarber
	}
	if err != nil {
		return st.Barber{}, e.Wrap("can't get barber", err)
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
		return "2000-01-01", st.ErrNoSavedWorkdates
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
	q = `CREATE TABLE IF NOT EXISTS appointments (
		id INTEGER PRIMARY KEY, 
		user_id INTEGER NOT NULL, 
		barber_id INTEGER NOT NULL, 
		service_id INTEGER, 
		date TEXT NOT NULL, 
		time TEXT NOT NULL,
		duration TEXT NOT NULL,
		cteated_at TEXT NOT NULL,
		UNIQUE (barber_id, date, time),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (barber_id) REFERENCES barbers(id) ON DELETE RESTRICT,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
		)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = `CREATE TABLE IF NOT EXISTS schedule (
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
	if barber.State.Valid && barber.Expiration.Valid {
		query = append(query, "state = ? , state_expiration = ?")
		args = append(args, barber.State, barber.Expiration)
	}
	args = append(args, barber.ID)
	q := fmt.Sprintf(`UPDATE barbers SET %s WHERE id = ?`, strings.Join(query, ", "))
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	_, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT) {
			err = st.ErrNonUniqueData
		}
		return e.Wrap("can't save barber", err)
	}
	return nil
}
