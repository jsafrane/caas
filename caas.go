package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

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

func get(name string) (Counter, error) {
	var counter Counter
	m := map[string]interface{}{}
	cql := "SELECT name, value FROM counter WHERE name=? LIMIT 1"
	query := session.Query(cql, name).Consistency(gocql.One)
	if err := query.MapScan(m); err != nil {
		return Counter{}, err
	}
	counter = Counter{
		Name:  m["name"].(string),
		Value: m["value"].(int64),
	}
	return counter, nil
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["counter_name"]
	log.Printf("Processing request for counter %s", name)

	err := bump(name)
	if err != nil {
		log.Printf("Error bumping counter %s: %s", name, err)
		w.WriteHeader(http.StatusInternalServerError)
		errJson := fmt.Sprintf("{\"error\": \"%s\"}\n", err)
		w.Write([]byte(errJson))
		return
	}
	counter, err := get(name)
	if err != nil {
		log.Printf("Error loading counter %s: %v", name, err)
		w.WriteHeader(http.StatusInternalServerError)
		errJson := fmt.Sprintf("{\"error\": \"%s\"}\n", err)
		w.Write([]byte(errJson))
		return
	}

	log.Printf("Counter %s bumped to %+v", name, counter)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(counter)
}

func connect() error {
	addr := os.Getenv("CASSANDRA_ADDRESS")
	if addr == "" {
		return fmt.Errorf("CASSANDRA_ADDRESS must be set")
	}

	// We need to resolve all IPs of cassandra server and connect to them
	ips, err := net.LookupHost(addr)
	if err != nil {
		return fmt.Errorf("cannot resolve %s: %s", addr, err)
	}

	log.Printf("Resolved cassandra address %s to %+v", addr, ips)

	cluster := gocql.NewCluster(ips...)
	cluster.Keyspace = "caas"

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
