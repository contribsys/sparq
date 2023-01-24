package public

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/activitystreams"
	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
)

func getUser(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nick := mux.Vars(r)["nick"]

		userdata := map[string]interface{}{}
		err := s.DB().QueryRowx(`
			select * from accounts	
			inner join account_securities
			on accounts.Id = account_securities.AccountId
			where accounts.Nick = ?`, nick).MapScan(userdata)
		if err == sql.ErrNoRows {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		ctype := r.Header.Get("Accept")
		// API call, render JSON for user
		if ctype == "application/activity+json" {
			url := "https://" + db.InstanceHostname + "/users/" + nick
			me := activitystreams.NewPerson(url)
			me.URL = url
			me.Name = userdata["FullName"].(string)
			me.PreferredUsername = userdata["Nick"].(string)
			me.AddPubKey(string(userdata["PublicKey"].([]uint8)))

			data, err := json.Marshal(me)
			if err != nil {
				httpError(w, err, http.StatusInternalServerError)
				return
			}
			w.Header().Add("Content-Type", "application/activity+json")
			_, _ = w.Write(data)
			return
		}

		// render html homepage for user
	}
}
