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

	_ "github.com/ziutek/mymysql/godrv"
)

type Env struct {
	DB models.Datastore
}

func internalServerError(w http.ResponseWriter, err error) {
	errMsg := "Internal Server Error"
	log.Println(errMsg + ": " + err.Error())
	http.Error(w, errMsg, http.StatusInternalServerError)
}

func (env *Env) PostUserHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &models.User{}
	if err := parseJSON(w, r.Body, reqUser); err != nil {
		return
	}

	if reqUser.Name == "" || reqUser.Email == "" || reqUser.Password == "" {
		errMsg := "Request body is missing field(s)"
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
	location := fmt.Sprintf("%s/%d", r.URL.Path, userID)
	w.Header().Add("Location", location)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqUser)
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
