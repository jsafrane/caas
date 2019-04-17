package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Counter struct {
	Name    string      `json:"name"`
	Value   int64       `json:"count"`
	Host    string      `json:"host"`
	DBStats []QueryStat `json:"dbStats""`
}

type QueryStat struct {
	Statement string `json:"statement"`
	Attempts  int    `json:"attempts"`
	Time      string `json:"time"`
	Host      string `json:"host"`
	Rows      int    `json:"rows"`
}

var db DB

func Get(w http.ResponseWriter, r *http.Request, renderer func(w http.ResponseWriter, counter Counter)) {
	vars := mux.Vars(r)
	name := vars["counter_name"]
	log.Printf("Processing request for counter %q", name)

	counter, err := db.IncrementAndGet(name)
	if err != nil {
		log.Printf("Error incrementing counter %q: %s", name, err)
		w.WriteHeader(http.StatusInternalServerError)
		errJson := fmt.Sprintf("{\"error\": \"%s\"}\n", err)
		w.Write([]byte(errJson))
		return
	}

	log.Printf("Counter %q bumped to %+v", name, counter)
	w.WriteHeader(http.StatusOK)

	renderer(w, counter)
}

func jsonRenderer(w http.ResponseWriter, counter Counter) {
	json.NewEncoder(w).Encode(counter)
}

func htmlRenderer(w http.ResponseWriter, counter Counter) {
	t := `
<html>
  <head><title>{{.Name}}</title></head>
<body>
<h1>Counter: {{.Name}}, value: {{.Value}}</h1>
<pre>
Web server: {{.Host}}
Queries:
{{range .DBStats}}
  DB server: {{.Host}}
  Query:     {{.Statement}}
  Attempts:  {{.Attempts}}
  Time:      {{.Time}}
{{end}}
</pre>
</body>
</html>`
	template.Must(template.New("html").Parse(t)).Execute(w, counter)
}

func GetJSON(w http.ResponseWriter, r *http.Request) {
	Get(w, r, jsonRenderer)
}

func GetHTML(w http.ResponseWriter, r *http.Request) {
	Get(w, r, htmlRenderer)
}

func main() {
	var err error
	db, err = NewCassandra()
	if err != nil {
		panic(err)
	}
	fmt.Println("cassandra init done")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/{counter_name}/json", GetJSON)
	router.HandleFunc("/{counter_name}/html", GetHTML)
	log.Fatal(http.ListenAndServe(":80", router))
}
