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

//
/* VARIABLES AND STRUCTS */
//

// Struct for db and db-related methods.
type Database struct {
	DB *sql.DB
}

// Struct for config variables.
type Config struct {
	Keys struct {
		JWT string `yaml:"jwt"`
	} `yaml:"keys"`
}

// Struct for banner parsing.
type BannerParse struct {
	Title       string `json:"title" binding:"required,min=2,max=128"` // 2-128 chars
	Description string `json:"description" binding:"max=256"`          // 0-256 chars
	Url         string `json:"url" binding:"required,min=1"`           // just put something
	ImageName   string `json:"image" binding:"required"`               // just upload something
}

// Struct for a full banner.
type Banner struct {
	BannerParse
	Author string `json:"author"` // login from jwt -> db -> author
}

// Variable for storing keys.
var Cfg = Config{}

//
/* VARIABLES AND STRUCTS END */
//

// Check if the keys are too short or too long.
func (cfg *Config) Validate() error {
	if tmp := len(cfg.Keys.JWT); tmp < 6 || tmp > 32 {
		return fmt.Errorf("the length of JWT key must be 6-32 characters (got %v) hint: change it in config/config.yaml", tmp)
	}
	return nil
}

// Load the database, creating one if it doesn't exist. Load config. note: don't forget to defer close.
func Init() (*Database, error) {
	data, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config (did you forget to create one?): %w", err)
	}

	if err = yaml.Unmarshal(data, &Cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := Cfg.Validate(); err != nil {
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
			url TEXT NOT NULL,
			image TEXT NOT NULL,
			time_created TEXT NOT NULL
		);
	`)
	var count int
	dbret := &Database{DB: db}

	if err := dbret.DB.QueryRow("SELECT COUNT(id) FROM banners").Scan(&count); err != nil {
		return nil, fmt.Errorf("failed to check the amount of banners: %w", err)
	}
	if count == 0 {
		if _, err = dbret.DB.Exec("INSERT INTO banners (title, description, author, url, image, time_created) VALUES (?, ?, ?, ?, ?, ?)", "Example", "", "*EXAMPLE*", "https://example.com", "placeholder.jpeg", time.Now().Format(time.DateTime)); err != nil {
			return nil, fmt.Errorf("failed to add example banner: %w", err)
		}
	}

	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return dbret, nil
}

// Check if user exists. Return false only if the user does not exist and no errors occurred.
func (db *Database) CheckUserExists(display, login string) error {
	var displayExists, loginExists bool

	err := db.DB.QueryRow(`
    SELECT
		EXISTS(SELECT 1 FROM users WHERE display = ?), 
		EXISTS(SELECT 1 FROM users WHERE login = ?)
	`, display, login).Scan(&displayExists, &loginExists)
	if err != nil {
		return fmt.Errorf("failed to check if user or display name exists")
	}

	switch {
	case displayExists:
		return fmt.Errorf("display name already exists")
	case loginExists:
		return fmt.Errorf("login already exists")
	default:
		return nil
	}
}

// Register a new user.
func (db *Database) Register(display, login string, password string) error {
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

	err := db.DB.QueryRow("SELECT password FROM users WHERE login = ?", login).Scan(&storedHash)
	if err != nil {
		return fmt.Errorf("user not exists")
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
func (db *Database) NewBanner(title, description, author, url, img string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	_, err := db.DB.Exec("INSERT INTO banners (title, description, author, url, image, time_created) VALUES (?, ?, ?, ?, ?, ?)", strings.TrimSpace(title), strings.TrimSpace(description), author, strings.TrimSpace(url), img, time.Now().Format(time.DateTime))
	if err != nil {
		return fmt.Errorf("failed to create a new banner 2")
	}
	return nil
}

// Update the amount of uploaded banners.
func (db *Database) UpdateBannerCount() (int, error) {
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

	rows, err := db.DB.Query("SELECT title, description, author, url, image FROM banners")
	if err != nil {
		return nil, fmt.Errorf("failed to update banner slice: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Banner
		err := rows.Scan(&b.Title, &b.Description, &b.Author, &b.Url, &b.ImageName)
		b.ImageName = "/static/img/banners/" + b.ImageName
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve row from db")
		}
		slice = append(slice, b)
	}
	return slice, nil
}

// Get the display name of a user by login.
func (db *Database) GetDisplay(login string) (string, error) {
	var display string

	err := db.DB.QueryRow("SELECT display FROM users WHERE login = ?", login).Scan(&display)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve display name from db")
	}
	return display, nil
}
