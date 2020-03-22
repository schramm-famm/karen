//handler checks the user that heimdall sends and then checks the email and password to verify it.

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"karen/models"
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
		mySQLErr, ok := err.(*mysql.MySQLError)
		if ok && mySQLErr.Number == 1065 {
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
	reqUser.ID = user.ID
	reqUser.Name = user.Name
	reqUser.Password = ""
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reqUser)
}

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

// PatchUserHandler updates specific columns of given user
func (env *Env) PatchUserHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &models.User{}
	if err := parseJSON(w, r.Body, reqUser); err != nil {
		return
	}
	if reqUser.Email == "" {
		errMsg := missingFieldMessage
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if reqUser.Name == "" && reqUser.Password == "" && reqUser.AvatarURL == "" {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reqUser)
		return
	}

	retUser, err := env.DB.UpdateUser(reqUser)
	if err != nil {
		mySQLErr, ok := err.(*mysql.MySQLError)
		println("code = " + string(mySQLErr.Number))
		if ok && mySQLErr.Number == 1065 {
			errMsg := fmt.Sprintf("User with email %s was not found", reqUser.Email)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusNotFound)
		} else {
			internalServerError(w, err)
		}
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(retUser)
}

// GetUserHandler gets a user returning specified columns
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
		responseUser.Password = ""
	} else {
		for _, column := range includes {
			switch column {
			case "id":
				responseUser.ID = user.ID
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
