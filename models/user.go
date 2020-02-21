package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"
)

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Salt      string `json:"salt"`
	AvatarURL string `json:"avatar_url"`
}

const (
	usersTable    string = "users"
	PW_SALT_BYTES        = 32
)

//Creates a user instance in the db and returns the user's id and the error
func (db *DB) CreateUser(user User) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}
	salt, err := generateSalt()
	if err != nil {
		return -1, err
	}
	unencryptedPasswordBytes := []byte(user.Password)
	encryptedPassword := string(encrypt(unencryptedPasswordBytes, salt))
	user.Password = encryptedPassword
	user.Salt = salt
	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(Name, Email, Password, Salt) ", usersTable)
	fmt.Fprintf(&b, "VALUES(?, ?, ?, ?)")
	res, err := tx.Exec(b.String(), user.Name, user.Email, user.Password, user.Salt)
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
func generateSalt() (string, error) {
	salt := make([]byte, PW_SALT_BYTES)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Fatal(err)
	}
	return string(salt), err
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
