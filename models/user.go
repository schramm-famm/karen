package models

import (
	"database/sql"
	"errors"
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
	usersTable string = "users"
)

//Authenticates the user's credentials and returns the user
func (db *DB) CheckUser(user *User) (*User, error) {
	userFromDB := &User{}
	queryString := fmt.Sprintf("SELECT ID, Name, Password FROM %s WHERE EMAIL=?", usersTable)
	err := db.QueryRow(queryString, user.Email).Scan(&(userFromDB.ID),
		&(userFromDB.Name), &(userFromDB.Password))
	if err != nil {
		return nil, err
	}
	err = checkPasswordHash(user.Password, userFromDB.Password)
	if err != nil {
		return nil, errors.New("password incorrect")
	}
	return userFromDB, err
}

func (db *DB) UpdateUser(user *User) (*User, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", usersTable)
	if user.AvatarURL != "" {
		fmt.Fprintf(&b, "AvatarURL='%s'", user.AvatarURL)
	}
	if user.Name != "" {
		if user.AvatarURL != "" {
			fmt.Fprintf(&b, ", ")
		}
		fmt.Fprintf(&b, "Name='%s'", user.Name)
	}
	if user.Password != "" {
		if user.AvatarURL != "" || user.Name != "" {
			fmt.Fprintf(&b, ", ")
		}
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&b, "Password='%s'", hashedPassword)
	}
	fmt.Fprintf(&b, " where Email=?")

	_, err := db.Exec(b.String(), user.Email)
	if err != nil {
		return nil, err
	}
	return user, err
}

//Creates a user instance in the db and returns the user's id and the error
func (db *DB) CreateUser(user *User) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return -1, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(Name, Email, Password, AvatarURL) ", usersTable)
	fmt.Fprintf(&b, "VALUES(?, ?, ?, ?)")
	res, err := tx.Exec(b.String(), user.Name, user.Email, user.Password, user.AvatarURL)
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

//ReadUser returns one user from the database given userID
func (db *DB) ReadUser(userID int64) (*User, error) {

	user := &User{}
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE ID=?", usersTable)

	err := db.QueryRow(queryString, userID).Scan(&(user.ID), &(user.Name), &(user.Email), &(user.Password), &(user.AvatarURL))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	log.Printf(`Read 1 row from "%s"`, usersTable)
	return user, nil
}

/*
func (db *DB) DeleteUser(user User) error {

}
*/

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
