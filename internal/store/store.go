package store

import (
	"database/sql"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	_ "modernc.org/sqlite"
)

// Struct for db and db-related methods.
type Database struct {
	DB *sql.DB
}

// Struct for config variables.
type Config struct {
	Keys struct {
		PassHash string `yaml:"passhash"`
		JWT      string `yaml:"jwt"`
	} `yaml:"keys"`
}

// Check if the keys are too short or too long.
func (cfg *Config) validate() error {
	if tmp := len(cfg.Keys.PassHash); tmp < 6 || tmp > 128 {
		return fmt.Errorf("the length of every key should be 6-128 characters (got %v for the PassHash key) hint: change them in config/config.yaml", tmp)
	}
	if tmp := len(cfg.Keys.JWT); tmp < 6 || tmp > 128 {
		return fmt.Errorf("the length of every key should be 6-128 characters (got %v for the JWT key) hint: change them in config/config.yaml", tmp)
	}
	return nil
}

var Cfg = Config{}

// Load the database, creating one if it doesn't exist. Load config. note: don't forget to defer close.
func Init() (*Database, error) {
	data, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config (did you forget to create one?): %w", err)
	}

	if err = yaml.Unmarshal(data, &Cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := Cfg.validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", "./data/db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS banners(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,	
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			color TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Database{DB: db}, nil
}

// Check if user exists. Return false only if the user does not exist and no errors occurred.
func (db *Database) CheckUserExists(login string) (bool, error) {
	var exists bool

	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)", login).Scan(&exists)
	if err != nil {
		return true, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return true, nil
	}
	return false, nil
}

// Register a new user.
func (db *Database) Register(login string, password string) error {
	password += Cfg.Keys.PassHash

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to encrypt the password: %w", err)
	}
	_, err = db.DB.Exec("INSERT INTO users (login, password) VALUES (?, ?)", login, string(hashedPass))
	if err != nil {
		return fmt.Errorf("failed to store user in the database: %w", err)
	}

	return nil
}

// Check user password. Return nil on success.
func (db *Database) CheckPassword(login string, password string) error {
	var storedHash []byte
	password += Cfg.Keys.PassHash

	err := db.DB.QueryRow("SELECT password FROM users WHERE login = ?", login).Scan(&storedHash)
	if err != nil {
		return fmt.Errorf("failed to retrieve password from db")
	}

	err = bcrypt.CompareHashAndPassword(storedHash, []byte(password))
	if err != nil {
		return fmt.Errorf("failed to compare passwords")
	}
	return nil
}
