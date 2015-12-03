package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

type dataHandler struct{}

type DataPoint struct {
	Enters int `json:"enters"`
	Exits int `json:"exits"`
}

const (
	minutesPerDay = 1440
)

var (
	enters   = 0
	exits    = 0

	data   [minutesPerDay][2]int
)

func initDataHandler() {
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

func (_ dataHandler) GET(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("type") {
	case "timeline":
		day := r.FormValue("day")
		month := r.FormValue("month")
		year := r.FormValue("year")
		filename := path.Join("data", year, month, day + ".json")
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
	case "dates":
		type Date struct {
			Year string `json:"year"`
			Month string `json:"month"`
			Day string `json:"day"`
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
						Year: year.Name(),
						Month: month.Name(),
						Day: day.Name()[:len(day.Name())-5],
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
		var data = DataPoint{
			Enters: enters,
			Exits: exits,
		}

		if jsonData, err := json.Marshal(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Write(jsonData)
		}
	}
}

func (_ dataHandler) POST(w http.ResponseWriter, r *http.Request) {
	enters++
	w.WriteHeader(http.StatusOK)
}

func (_ dataHandler) DELETE(w http.ResponseWriter, r *http.Request) {
	exits++
	w.WriteHeader(http.StatusOK)
}

func updateCurrentMinuteData() {
	now := time.Now().Add(8 * time.Hour)
	minuteOfDay := now.Hour()*60 + now.Minute()
	data[minuteOfDay][0] = enters
	data[minuteOfDay][1] = exits

	// Save data
	saveJSON(path.Join("data", toDataFilename(time.Now())), data)
	log.Println("Saved data for minute", minuteOfDay)

	if minuteOfDay == 24 * 60 - 1 {
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
		strconv.Itoa(t.Day()) + ".json",
	)
}
