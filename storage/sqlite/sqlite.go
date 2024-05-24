package sqlite

import (
	"barbershop-bot/lib/e"
	"barbershop-bot/storage"
	"context"
	"database/sql"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Storage struct {
	db       *sql.DB
	location *time.Location
	mutex    *sync.Mutex
}

// New creates a new SQLite storage.
func New(path string, location *time.Location, mutex *sync.Mutex) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, e.Wrap("can't open database", err)
	}

	if err := db.Ping(); err != nil {
		return nil, e.Wrap("can't connect to database", err)
	}
	return &Storage{db: db, location: location, mutex: mutex}, nil
}

// Init prepares the storage for use. It creates the necessary tables if they do not exist
// and imports the barber ID from the environment variable into the storage
// if there is not a single barber in the storage.
func (s *Storage) Init(ctx context.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't initialize storage", err) }()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err = s.createTables(ctx)
	if err != nil {
		return err
	}

	check, err := s.checkBarbers(ctx)
	if err != nil {
		return err
	}
	if !check {
		return s.addBarberFromEnv(ctx)
	}
	return nil
}

// BarberIDs return a slice of IDs of all barbers.
func (s *Storage) BarberIDs(ctx context.Context) (barberIDs []int64, err error) {
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

// LatestWorkDate return the latest work date saved for barber with barberID.
// Result have a format of time.Time with HH:MM:SS set to 00:00:00.
func (s *Storage) LatestWorkDate(ctx context.Context, barberID int64) (time.Time, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q := `SELECT COUNT(*) FROM schedule WHERE barber_id = ?`

	var count int
	if err := s.db.QueryRowContext(ctx, q, barberID).Scan(&count); err != nil {
		return time.Time{}, e.Wrap("can't check if any work date exists", err)
	}
	if count == 0 {
		return time.Time{}, storage.ErrNoSavedWorkdates
	} else {
		q = `SELECT MAX(date) FROM schedule WHERE barber_id = ?`
		var date string
		if err := s.db.QueryRowContext(ctx, q, barberID).Scan(&date); err != nil {
			return time.Time{}, e.Wrap("can't get latest workdate", err)
		}
		return time.ParseInLocation(time.DateOnly, date, s.location)
	}
}

// SaveWorkday saves Workday type value to storage
func (s *Storage) SaveWorkday(ctx context.Context, workday storage.Workday) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q := `INSERT INTO schedule (barber_id, date, start_time, end_time) VALUES (?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, q, workday.BarberID, workday.Date.Format(time.DateOnly), workday.StartTime, workday.EndTime)
	if err != nil {
		return e.Wrap("can't save workday", err)
	}
	return nil
}

// SaveBarberName saves name for barber with barberID
func (s *Storage) SaveBarberName(ctx context.Context, name string, barberID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q := `UPDATE barbers SET name = ? WHERE id = ?`

	_, err := s.db.ExecContext(ctx, q, name, barberID)
	if err != nil {
		return e.Wrap("can't save barber name", err)
	}
	return nil
}

// SaveBarberState saves state of dialog for barber with barberID.
// It also saves expiration time for this state taking it as 24 hours after SaveBarberState call.
func (s *Storage) SaveBarberState(ctx context.Context, state uint8, barberID int64) error {
	expiration := time.Now().In(s.location).Add(24 * time.Hour).Format(time.DateTime)
	q := `UPDATE barbers SET state = ? , state_expiration = ? WHERE id = ?`

	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, err := s.db.ExecContext(ctx, q, state, expiration, barberID)
	if err != nil {
		return e.Wrap("can't save state of dialog with barber", err)
	}
	return nil
}

func (s *Storage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Close()
}

// checkBarbers checks if there is at least one saved barber. It returns true if there is at least one saved barber.
// Othewise it returns false. checkBarbers should be protected with mutex externally.
func (s *Storage) checkBarbers(ctx context.Context) (bool, error) {
	q := `SELECT COUNT(*) FROM barbers`

	var count int

	if err := s.db.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return false, e.Wrap("can't check barbers", err)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// addBarberFromEnv gets Barber ID from environment variable and saves it.
// addBarberFromEnv should be protected with mutex externally.
func (s *Storage) addBarberFromEnv(ctx context.Context) error {
	barberID, err := strconv.ParseInt(os.Getenv("BarberID"), 10, 64)
	if err != nil {
		return e.Wrap("can't get barberID from environment variable", err)
	}
	q := `INSERT INTO barbers (id) VALUES (?)`

	if _, err := s.db.ExecContext(ctx, q, barberID); err != nil {
		return e.Wrap("can't save barberID", err)
	}
	return nil
}

// createTables create tables if not exists.
func (s *Storage) createTables(ctx context.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't create table", err) }()

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
