package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var generateRandomData = flag.Bool("generate-random-data", false, "Generates random data in JSON format")

func dataEndpoint(w http.ResponseWriter, r *http.Request) {
	var handler dataHandler

	switch r.Method {
	case "GET":
		handler.GET(w, r)
	case "POST":
		handler.POST(w, r)
	case "DELETE":
		handler.DELETE(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func bodiesEndpoint(w http.ResponseWriter, r *http.Request) {
	var handler bodiesHandler

	switch r.Method {
	case "GET":
		handler.GET(w, r)
	case "POST":
		handler.POST(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	flag.Parse()

	if *generateRandomData {
		doGenerateRandomData()
		return
	}

	initDataHandler()

	router := mux.NewRouter()
	router.HandleFunc("/data", dataEndpoint)
	router.HandleFunc("/bodies", bodiesEndpoint)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}

func doGenerateRandomData() {
	var dps [minutesPerDay][2]int
	enters := 0
	exits := 0

	rand.Seed(time.Now().Unix())
	for i := range dps {
		enters += rand.Int() % 4
		exits += rand.Int() % 3
		dps[i][0] = enters
		dps[i][1] = exits
	}

	jsonData, err := json.Marshal(dps)
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(jsonData)
}
