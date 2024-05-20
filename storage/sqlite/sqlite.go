package sqlite

import (
	"barbershop-bot/lib/e"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Storage struct {
	db *sql.DB
}

// New creates a new SQLite storage.
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}
	return &Storage{db: db}, nil
}

// Init prepares the storage for use. It creates the necessary tables if they do not exist
// and imports the barber ID from the environment variable into the storage
// if there is not a single barber in the storage.
func (s *Storage) Init(ctx context.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't initialize storage", err) }()

	err = s.createTables(ctx)
	if err != nil {
		return err
	}

	err = s.checkBarbers(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// BarberIDs return a slice of IDs of all barbers.
func (s *Storage) BarberIDs(ctx context.Context) (barberIDs []int64, err error) {
	defer func() { err = e.WrapIfErr("can't get barber IDs", err) }()

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

func (s *Storage) checkBarbers(ctx context.Context) error {
	q := `SELECT COUNT(*) FROM barbers`

	var count int

	if err := s.db.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return fmt.Errorf("can't check barbers: %w", err)
	}
	if count == 0 {
		return s.addBarberFromEnv(ctx)
	}
	return nil
}

func (s *Storage) addBarberFromEnv(ctx context.Context) error {
	barberID, err := strconv.ParseInt(os.Getenv("BarberID"), 10, 64)
	if err != nil {
		return fmt.Errorf("wrong barberID in environment variable: %w", err)
	}
	q := `INSERT INTO barbers (id) VALUES (?)`

	if _, err := s.db.ExecContext(ctx, q, barberID); err != nil {
		return fmt.Errorf("can't add barber from envenvironment: %w", err)
	}
	return nil
}

func (s *Storage) createTables(ctx context.Context) (err error) {
	defer func() { err = e.WrapIfErr("can't create table", err) }()

	q := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY, 
		name TEXT, 
		phone TEXT, 
		chat_id INTEGER UNIQUE, 
		state TEXT DEFAULT "StateStart",
		state_expiration TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}

	q = `CREATE TABLE IF NOT EXISTS barbers (
		id INTEGER PRIMARY KEY, 
		name TEXT UNIQUE, 
		bio TEXT UNIQUE, 
		phone TEXT UNIQUE, 
		chat_id INTEGER UNIQUE, 
		state TEXT DEFAULT "StateStart",
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
