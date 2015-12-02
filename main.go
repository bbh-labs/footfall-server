package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

type DataPoint struct {
	Enters int `json:"enters"`
	Exits  int `json:"exits"`
}

const (
	minutesPerDay = 1440
)

var (
	enters    = 0
	exits     = 0
	data      [minutesPerDay][2]int
	bodiesMap = make(map[string]string)
	generateData = flag.Bool("generate-data", false, "Generates data in JSON format")
)

func updateCurrentMinuteData() {
	now := time.Now().Add(8 * time.Hour)
	minuteOfDay := now.Hour()*60 + now.Minute()
	data[minuteOfDay][0] = enters
	data[minuteOfDay][1] = exits

	// Save data
	saveJSON(path.Join("data", toDataFilename(time.Now())), data)
	log.Println("Saved data for minute", minuteOfDay)

	if minuteOfDay == minutesPerDay-1 {
		// Clear data
		for i := range data {
			for j := range data[i] {
				data[i][j] = 0
			}
		}
	}
}

func toDataFilename(t time.Time) string {
	return path.Join(
		strconv.Itoa(t.Year()),
		strconv.Itoa(int(t.Month())),
		strconv.Itoa(t.Day())+".json",
	)
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		day := r.FormValue("day")
		month := r.FormValue("month")
		year := r.FormValue("year")
		filename := path.Join("data", year, month, day+".json")
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				jsonData, err := json.Marshal(data)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Write(jsonData)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			http.ServeFile(w, r, filename)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func datesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		type Date struct {
			Year  string `json:"year"`
			Month string `json:"month"`
			Day   string `json:"day"`
		}
		var dates []Date

		years, err := ioutil.ReadDir("data")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, year := range years {
			months, err := ioutil.ReadDir(path.Join("data", year.Name()))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for _, month := range months {
				days, err := ioutil.ReadDir(path.Join("data", year.Name(), month.Name()))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				for _, day := range days {
					dayName := day.Name()
					if len(dayName) < 5 {
						continue
					}
					dates = append(dates, Date{
						Year:  year.Name(),
						Month: month.Name(),
						Day:   day.Name()[:len(day.Name())-5],
					})
				}
			}
		}

		if jsonData, err := json.Marshal(dates); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Write(jsonData)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func visitHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var data = DataPoint{
			Enters: enters,
			Exits:  exits,
		}

		if jsonData, err := json.Marshal(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Write(jsonData)
		}
	case "POST":
		enters++
		w.WriteHeader(http.StatusOK)
	case "DELETE":
		exits++
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func bodiesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if data, err := json.Marshal(bodiesMap); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write([]byte(data))
		}
	case "POST":
		location := r.FormValue("location")
		bodiesMap[location] = r.FormValue("bodies")
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func init() {
	flag.Parse()

	if *generateData {
		doGenerateData()
		os.Exit(0)
	}

	now := time.Now().Add(8 * time.Hour)
	if err := loadJSON(path.Join("data", toDataFilename(now)), &data); err != nil {
		if os.IsNotExist(err) {
			// Ignore 'not exist' errors
		} else {
			log.Println(err)
		}
	} else {
		log.Println("Loaded data for", now.Weekday(), now.Day(), now.Month(), now.Year())
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				updateCurrentMinuteData()
			}
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)
		<-c

		// Save data
		now := time.Now().Add(8 * time.Hour)
		saveJSON(path.Join("data", toDataFilename(now)), data)
		log.Println("Saved data for", now.Weekday(), now.Day(), now.Month(), now.Year())
		os.Exit(0)
	}()
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

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/visit", visitHandler)
	router.HandleFunc("/bodies", bodiesHandler)
	router.HandleFunc("/dates", datesHandler)
	router.HandleFunc("/timeline", timelineHandler)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
