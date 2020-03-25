package models

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	// MySQL database driver
	_ "github.com/go-sql-driver/mysql"
)

// Datastore defines the CRUD operations of models in the database.
type Datastore interface {
	CreateUser(user *User) (int64, error)
	CheckUser(user *User) (*User, error)
	ReadUser(userID int64) (*User, error)
	UpdateUser(user *User) (int64, error)
	DeleteUser(int64) (int64, error)
}

// DB represents an SQL database connection.
type DB struct {
	*sql.DB
}

// NewDB initializes a new DB.
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	if err = setupDB(db); err != nil {
		return nil, err
	}
	db.Close()

	// Re-open, this time with a specified database
	slashIndex := strings.Index(dataSourceName, "/")
	if slashIndex < 0 {
		return nil, fmt.Errorf("Invalid data source name")
	}
	db, err = sql.Open(
		"mysql",
		dataSourceName[:slashIndex+1]+"karen"+dataSourceName[slashIndex+1:],
	)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// setupDB creates the necessary "karen" database and "users" table if they
// don't already exist and makes the db connection use the "karen" database.
func setupDB(db *sql.DB) error {
	sqlScript, err := ioutil.ReadFile("dbSchema.sql")
	if err != nil {
		return err
	}

	statements := strings.Split(string(sqlScript), ";")
	if len(statements) > 0 {
		statements = statements[:len(statements)-1]
	}

	for _, statement := range statements {
		_, err = db.Exec(statement)
		if err != nil {
			return err
		}
	}
	return nil
}
