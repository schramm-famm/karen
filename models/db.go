package models

import (
	"database/sql"

	// MySQL database driver
	_ "github.com/go-sql-driver/mysql"
)

type Datastore interface {
	CreateUser(user *User) (int64, error)
	CheckUser(user *User) (*User, error)
}

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
