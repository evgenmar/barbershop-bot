package sqlite

import (
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"context"
	"database/sql"
	"sync"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Storage struct {
	db    *sql.DB
	mutex *sync.Mutex
}

// New creates a new SQLite storage.
func New(path string, mutex *sync.Mutex) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, e.Wrap("can't open database", err)
	}

	if err := db.Ping(); err != nil {
		return nil, e.Wrap("can't connect to database", err)
	}
	return &Storage{db: db, mutex: mutex}, nil
}

func (s *Storage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Close()
}

// CreateBarberID saves new BarberID to storage
func (s *Storage) CreateBarberID(ctx context.Context, barberID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	q := `INSERT INTO barbers (id) VALUES (?)`

	if _, err := s.db.ExecContext(ctx, q, barberID); err != nil {
		return e.Wrap("can't save barberID", err)
	}
	return nil
}

// CreateWorkday saves new Workday to storage
func (s *Storage) CreateWorkday(ctx context.Context, workday storage.Workday) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q := `INSERT INTO schedule (barber_id, date, start_time, end_time) VALUES (?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, q, workday.BarberID, workday.Date, workday.StartTime, workday.EndTime)
	if err != nil {
		return e.Wrap("can't save workday", err)
	}
	return nil
}

// FindAllBarberIDs return a slice of IDs of all barbers.
func (s *Storage) FindAllBarberIDs(ctx context.Context) (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

// GetLatestWorkDate returns the latest work date saved for barber with barberID.
// If there is no saved work dates it returns ("2000-01-01", ErrNoSavedWorkdates).
func (s *Storage) GetLatestWorkDate(ctx context.Context, barberID int64) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY, 
		name TEXT, 
		phone TEXT, 
		chat_id INTEGER UNIQUE, 
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
		chat_id INTEGER UNIQUE, 
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

// UpdateBarberName saves new name for barber with barberID
func (s *Storage) UpdateBarberName(ctx context.Context, name string, barberID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q := `UPDATE barbers SET name = ? WHERE id = ?`

	_, err := s.db.ExecContext(ctx, q, name, barberID)
	if err != nil {
		return e.Wrap("can't save barber name", err)
	}
	return nil
}

// UpdateBarberStatus saves new status of dialog for barber with barberID.
func (s *Storage) UpdateBarberStatus(ctx context.Context, status storage.Status, barberID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	q := `UPDATE barbers SET state = ? , state_expiration = ? WHERE id = ?`

	_, err := s.db.ExecContext(ctx, q, status.State, status.Expiration, barberID)
	if err != nil {
		return e.Wrap("can't save state of dialog with barber", err)
	}
	return nil
}
