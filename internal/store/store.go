package store

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

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

// Struct for banner parsing.
type BannerParse struct {
	Title       string `json:"title" binding:"required,min=2,max=128"`
	Description string `json:"description" binding:"max=256"`
	Url         string `json:"url" binding:"required,min=1"`
}

// Struct for a full banner.
type Banner struct {
	BannerParse
	Author string `json:"author"`
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
			display TEXT NOT NULL UNIQUE,
			login TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS banners(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL UNIQUE,
			description TEXT,
			author TEXT NOT NULL,
			url TEXT,
			time_created TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Database{DB: db}, nil
}

// Check if user exists. Return false only if the user does not exist and no errors occurred.
func (db *Database) CheckUserExists(display, login string) error {
	var exists bool

	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE display = ?)", display).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if user exists")
	}
	if exists {
		return fmt.Errorf("display name already exists")
	}

	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)", login).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if user exists")
	}
	if exists {
		return fmt.Errorf("login already exists")
	}

	return nil
}

// Register a new user.
func (db *Database) Register(display, login string, password string) error {
	password += Cfg.Keys.PassHash

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to encrypt the password")
	}
	_, err = db.DB.Exec("INSERT INTO users (display, login, password) VALUES (?, ?, ?)", display, login, string(hashedPass))
	if err != nil {
		return fmt.Errorf("failed to store user in the database")
	}

	return nil
}

// Check user password. Return nil on success.
func (db *Database) CheckPassword(login, password string) error {
	var storedHash []byte
	password += Cfg.Keys.PassHash

	err := db.DB.QueryRow("SELECT password FROM users WHERE login = ?", login).Scan(&storedHash)
	if err != nil {
		return fmt.Errorf("failed to retrieve password from db")
	}

	err = bcrypt.CompareHashAndPassword(storedHash, []byte(password))
	if err != nil {
		return fmt.Errorf("failed to login")
	}
	return nil
}

// Check if a banner already exists. Return false only if the banner does not exist and no errors occurred.
func (db *Database) CheckBannerExists(title string) error {
	var exists bool

	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM banners WHERE title = ?)", title).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if banner exists")
	}
	if exists {
		return fmt.Errorf("banner with this title already exists")
	}

	return nil
}

// Store a banner in the database.
func (db *Database) NewBanner(title, description, author, url string) error {
	err := db.DB.QueryRow("SELECT display FROM users WHERE login = ?", author).Scan(&author)
	if err != nil {
		return fmt.Errorf("failed to create a new banner 1")
	}
	_, err = db.DB.Exec("INSERT INTO banners (title, description, author, url, time_created) VALUES (?, ?, ?, ?, ?)", strings.TrimSpace(title), strings.TrimSpace(description), author, strings.TrimSpace(url), time.Now().Format(time.DateTime))
	if err != nil {
		return fmt.Errorf("failed to create a new banner 2")
	}
	return nil
}

// Update the amount of uploaded banners.
func (db *Database) UpdateGames() (int, error) {
	var count int
	err := db.DB.QueryRow("SELECT COUNT(id) FROM banners").Scan(&count)
	if err != nil {
		return -1, fmt.Errorf("failed to update game counter: %w", err)
	}
	return count, nil
}

// Update banner slice.
func (db *Database) UpdateBannerSlice() ([]Banner, error) {
	var slice []Banner

	rows, err := db.DB.Query("SELECT title, description, author, url FROM banners")
	if err != nil {
		return nil, fmt.Errorf("failed to update banner slice: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Banner
		err := rows.Scan(&b.Title, &b.Description, &b.Author, &b.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve row from db")
		}
		slice = append(slice, b)
	}
	return slice, nil
}
