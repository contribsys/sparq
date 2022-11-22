package public

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/contribsys/sparq/activitystreams"
	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("User not found")
)

func AddPublicEndpoints(mux *mux.Router) {
	mux.HandleFunc("/users/{nick:[a-z0-9]{4,16}}", getUser)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	nick := mux.Vars(r)["nick"]

	userdata := map[string]interface{}{}
	err := db.Database().QueryRowx(`
	select *
	from users
	inner join user_securities
	on users.id = user_securities.userid
	where users.nick = ?`, nick).MapScan(userdata)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := "https://" + db.InstanceHostname + "/users/" + nick
	me := activitystreams.NewPerson(url)
	me.URL = url
	me.Name = userdata["FullName"].(string)
	me.PreferredUsername = userdata["Nick"].(string)
	me.AddPubKey(string(userdata["PublicKey"].([]uint8)))

	data, err := json.Marshal(me)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/activity+json")
	_, _ = w.Write(data)
}
