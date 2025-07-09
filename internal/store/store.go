package store

import (
	"database/sql"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	_ "github.com/mattn/go-sqlite3"
)

// Struct for db and db-related methods.
type Database struct {
	DB *sql.DB
}

// Struct for config variables
type Config struct {
	DB struct {
		Secret string `yaml:"secret"`
	} `yaml:"db"`
}

// Check if the password is too short.
func (cfg *Config) validate() error {
	if tmp := len(cfg.DB.Secret); tmp < 6 || tmp > 128 {
		return fmt.Errorf("secret length should be 6-128 characters (got %v) hint: change it in config/config.yaml", tmp)
	}
	return nil
}

var config = Config{}

// Load the database, creating one if it doesn't exist. Load config. note: don't forget to defer close.
func Init() (*Database, error) {
	data, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config (did you forget to create one?): %w", err)
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", "./data/db")
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
	password += config.DB.Secret

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

// Check user password. Error if user does not exist.
func (db *Database) CheckPassword(login string, password string) error {
	var storedHash []byte
	password += config.DB.Secret

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
