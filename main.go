package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron"
)

var outages Response

func UpdateSchedule() {
	outages = ScrapeOutages()

	fmt.Print("Updated outage data at: ")
	fmt.Println(time.Now())
}

func GetOutages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonString, _ := json.Marshal(outages)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func main() {
	// Create new schedule at startup
	UpdateSchedule()

	// Schedule update every hour
	c := cron.New()
	c.AddFunc("@every 5m", UpdateSchedule)
	c.Start()

	router := httprouter.New()

	// Root api call
	router.GET("/api/", GetOutages)

	// Home page
	router.NotFound = http.FileServer(http.Dir("./static"))

	var port = os.Getenv("PORT")

	if port == "" {
		port = "9000"
	}

	http.ListenAndServe(":"+port, router)
}
