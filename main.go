package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/cloudfoundry-community/go-cfenv"
)

func handler(w http.ResponseWriter, r *http.Request) {
	var service cfenv.Service

	appEnv, errCfenv := cfenv.Current()
	if errCfenv == nil {
		for _, mappedServices := range appEnv.Services {
			for _, s := range mappedServices {
				service = s
				break
			}
		}
	}

	vcapServices := os.Getenv("VCAP_SERVICES")

	errDbMsg := ""

	values := []int{}

	uri, _ := service.Credentials["uri"].(string)
	db, errDb := sql.Open("postgres", uri+"?sslmode=disable")
	if errDb == nil {
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS entries (id int)")
		if err != nil {
			errDbMsg += fmt.Sprintf("CREATE TABLE ERROR %v\n", err)
		}
		_, err = db.Exec(fmt.Sprintf("INSERT INTO entries (id) VALUES (%d)", rand.Intn(1000)))
		if err != nil {
			errDbMsg += fmt.Sprintf("INSERT ERROR %v\n", err)
		}
		rows, err := db.Query("SELECT id FROM entries")
		if err != nil {
			errDbMsg += fmt.Sprintf("SELECT ERROR %v\n", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				errDbMsg += fmt.Sprintf("row scan ERROR %v\n", err)
			} else {
				values = append(values, id)
			}
		}
	}

	fmt.Fprintf(w, fmt.Sprintf("errCfenv = %v\nerrDb = %v\nerrDbMsg = %v\nenv VCAP_SERVICES: %s\ncredentials = %v\nvalues = %v", errCfenv, errDb, errDbMsg, vcapServices, service.Credentials, values))
}

func main() {
	port := os.Getenv("PORT")
	if len(port) < 1 {
		port = "8080"
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}
