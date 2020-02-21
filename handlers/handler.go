//handler checks the user that heimdall sends and then checks the email and password to verify it.

package handlers

import (
	"karen/models"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/ziutek/mymysql/godrv"
)

var (
	rc             *http.Client
	karenAuth      = "/karen/api/auth"
	privateKeyPath = "id_rsa"
	whitelist      = []string{"/karen/api/auth", "/karen"}
)

func init() {
	rc = &http.Client{
		Timeout: time.Second * 10,
	}
}

func internalServerError(w http.ResponseWriter, err error) {
	errMsg := "Internal Server Error"
	log.Println(errMsg + ": " + err.Error())
	http.Error(w, errMsg, http.StatusInternalServerError)
}

func (env) PostUserHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &models.User{}
	if err := parseJSON(w, r.Body, reqUser); err != nil{
		return
	}

	if reqUser.Name == "" || reqUser.Email == "" || reqUser.Password == ""
	{
		errMsg := "Request body is missing field(s)"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	userID, err := env.DB.CreateUser(reqUser)
	if err != nil {
		internalServerError(w, err)
		return
	}

	reqUser.ID = userID
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
