//handler checks the user that heimdall sends and then checks the email and password to verify it.

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"karen/models"
	"karen/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/ziutek/mymysql/godrv"

	// MySQL database driver
	"github.com/go-sql-driver/mysql"
)

const (
	missingFieldMessage string = "Request body is missing field(s)"
)

type Env struct {
	DB models.Datastore
}

func internalServerError(w http.ResponseWriter, err error) {
	errMsg := "Internal Server Error"
	log.Println(errMsg + ": " + err.Error())
	http.Error(w, errMsg, http.StatusInternalServerError)
}

// PostAuthHandler verifies provided credentials against the database.
func (env *Env) PostAuthHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &models.User{}
	if err := parseJSON(w, r.Body, reqUser); err != nil {
		return
	}
	if reqUser.Email == "" || reqUser.Password == "" {
		errMsg := missingFieldMessage
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	user, err := env.DB.CheckUser(reqUser)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := fmt.Sprintf("User with email %s was not found", reqUser.Email)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusNotFound)
		} else if err.Error() == "password incorrect" {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			internalServerError(w, err)
		}
		return
	}
	user.Password = ""
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// PostUserHandler creates a single new user.
func (env *Env) PostUserHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &models.User{}
	if err := parseJSON(w, r.Body, reqUser); err != nil {
		return
	}

	if reqUser.Name == "" || reqUser.Email == "" || reqUser.Password == "" {
		errMsg := missingFieldMessage
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if reqUser.AvatarURL == nil {
		reqUser.AvatarURL = utils.StringPtr("")
	}

	userID, err := env.DB.CreateUser(reqUser)
	if err != nil {
		mySQLErr, ok := err.(*mysql.MySQLError)
		if ok && mySQLErr.Number == 1062 {
			errMsg := fmt.Sprintf("User already exists with email %s", reqUser.Email)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusConflict)
		} else {
			internalServerError(w, err)
		}
		return
	}

	reqUser.ID = userID
	reqUser.Password = ""
	location := fmt.Sprintf("%s/self", r.URL.Path)
	w.Header().Add("Location", location)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqUser)
}

// DeleteUserHandler removes a single user.
func (env *Env) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil || userID <= 0 {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	rowsAffected, err := env.DB.DeleteUser(userID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if rowsAffected == 0 {
		errMsg := "User not found"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PatchUserHandler updates a single user.
func (env *Env) PatchUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	reqUser := &models.User{}
	if err := parseJSON(w, r.Body, reqUser); err != nil {
		return
	}
	reqUser.ID = userID
	if reqUser.Name == "" && reqUser.Email == "" && reqUser.Password == "" && reqUser.AvatarURL == nil {
		errMsg := `Request body must have one of "email", "name", "password", or "avatar_url"`
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	dbUser, err := env.DB.ReadUser(userID)
	if dbUser == nil {
		errMsg := "User not found"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}
	if err != nil {
		internalServerError(w, err)
		return
	}

	newUser := dbUser.Merge(reqUser)

	rowsAffected, err := env.DB.UpdateUser(newUser)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if rowsAffected == 0 {
		errMsg := "User not found"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	newUser.ID = 0
	newUser.Password = ""
	json.NewEncoder(w).Encode(newUser)
}

// GetUserHandler gets a single user, returning specified columns.
func (env *Env) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var userID int64
	var err error
	if vars["user-id"] != "" {
		userID, err = strconv.ParseInt(vars["user-id"], 10, 64)
	} else {
		userID, err = strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	}
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	user, err := env.DB.ReadUser(userID)
	if user == nil {
		errMsg := "User not found"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}
	if err != nil {
		internalServerError(w, err)
		return
	}

	responseUser := &models.User{}
	r.ParseForm()
	includes := r.Form["includes"]
	if includes == nil {
		responseUser = user
		responseUser.ID = 0
		responseUser.Password = ""
	} else {
		for _, column := range includes {
			switch column {
			case "name":
				responseUser.Name = user.Name
			case "email":
				responseUser.Email = user.Email
			case "avatar_url":
				responseUser.AvatarURL = user.AvatarURL
			default:
				errMsg := "Invalid includes format"
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}
		}
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseUser)
}

func parseJSON(w http.ResponseWriter, body io.ReadCloser, bodyObj interface{}) error {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		errMsg := "Failed to read request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	if err := json.Unmarshal(bodyBytes, bodyObj); err != nil {
		errMsg := "Failed to parse request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	return nil
}
