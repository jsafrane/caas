package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
)

type Counter struct {
	Name  string `json:"name"`
	Value int64  `json:"count"`
}

var session *gocql.Session

func bump(name string) error {
	return session.Query(`UPDATE counter SET value=value+1 WHERE name = ?`, name).Exec()
}

func get(name string) (Counter, int) {
	var counter Counter
	var found bool
	m := map[string]interface{}{}
	query := "SELECT name, value FROM counter WHERE name=? LIMIT 1"
	iterable := session.Query(query, name).Consistency(gocql.One).Iter()
	for iterable.MapScan(m) {
		found = true
		counter = Counter{
			Name:  m["name"].(string),
			Value: m["value"].(int64),
		}
	}
	if !found {
		return Counter{}, http.StatusNotFound
	}
	return counter, http.StatusOK
}

func Get(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["counter_name"]

	err := bump(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	counter, errCode := get(name)
	w.WriteHeader(errCode)
	_ = json.NewEncoder(w).Encode(counter)
}

func connect() error {
	addr := os.Getenv("CASSANDRA_ADDRESS")
	if addr == "" {
		return fmt.Errorf("CASSANDRA_ADDRESS must be set")
	}
	cluster := gocql.NewCluster(addr)
	cluster.Keyspace = "caas"
	var err error
	session, err = cluster.CreateSession()
	return err
}

func main() {
	err := connect()
	if err != nil {
		panic(err)
	}
	fmt.Println("cassandra init done")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/{counter_name}/json", Get)
	log.Fatal(http.ListenAndServe(":80", router))
}
