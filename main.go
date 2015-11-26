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

var generateData = flag.Bool("generate-data", false, "Generates data in JSON format")

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

	if *generateData {
		doGenerateData()
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

func doGenerateData() {
	var hourlyActivities = [24]int{
		0, // 12AM
		0, // 1AM
		0, // 2AM
		0, // 3AM
		0, // 4AM
		0, // 5AM
		0, // 6AM
		1, // 7AM
		2, // 8AM
		9, // 9AM
		4, // 10AM
		2, // 11AM
		5, // 12PM
		8, // 1PM
		4, // 2PM
		3, // 3PM
		2, // 4PM
		2, // 5PM
		1, // 6PM
		2, // 7PM
		2, // 8PM
		1, // 9PM
		0, // 10PM
		0, // 11PM
	}

	rand.Seed(time.Now().Unix())

	var dps [minutesPerDay][2]int
	enters := 0
	exits := 0
	for minute := range dps {
		hour := minute / 60
		enters += rand.Int() % 3 * hourlyActivities[hour]
		exits += rand.Int() % 3 * hourlyActivities[hour]
		dps[minute][0] = enters
		dps[minute][1] = exits
	}

	jsonData, err := json.Marshal(dps)
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(jsonData)
}
