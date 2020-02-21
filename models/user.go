package models

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

const (
	usersTable    string = "users"
	PW_SALT_BYTES        = 32
)

//Creates a user instance in the db and returns the user's id and the error
func (db *DB) CreateUser(user *User) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}

	user.Password, err = HashPassword(user.Password)
	if err != nil {
		return -1, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(Name, Email, Password) ", usersTable)
	fmt.Fprintf(&b, "VALUES(?, ?, ?)")
	res, err := tx.Exec(b.String(), user.Name, user.Email, user.Password)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Created %d row(s) in "%s"`, rowCount, usersTable)
	} else {
		tx.Rollback()
		return -1, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return userID, err
}

/*
func (db *DB) ReadUser(user User) ([]*User, error) {

}

func (db *DB) UpdateUser(user User) ([]*User, error) {

}

func (db *DB) DeleteUser(user User) error {

}

func (db *DB) CheckUser(email String, password String) ([]*User, error) {

}
*/
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
