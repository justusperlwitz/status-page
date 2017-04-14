package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/negroni"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type SiteStatus struct {
	SiteName   string
	SiteUptime int
	SiteURL    string
	Ping       float64
	Status     string
}

const (
	DefaultPort = "8080"
)

var (
	siteStatus   = &SiteStatus{}
	updateTicker = time.NewTicker(5 * time.Second)
)

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	if err := t.Execute(w, siteStatus); err != nil {
		log.Printf("Error when rendering template: %+v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func measureLatency(URL string) float64 {
	start := time.Now()
	result, err := http.Get(URL)
	if err != nil {
		log.Panicf("Something went wrong: %+v", err)
	}
	defer result.Body.Close()
	elapsed := time.Since(start).Seconds()
	siteStatus.Status = result.Status
	return elapsed
}

func updateSiteStatus() {
	for {
		select {
		case <-updateTicker.C:
			siteStatus.Ping = measureLatency(siteStatus.SiteURL)
		}
	}
}

func init() {
	var present bool
	siteStatus.SiteName, present = os.LookupEnv("SITE_NAME")
	if !present {
		log.Panic("SITE_NAME must be defined in environment")
	}

	siteStatus.SiteURL, present = os.LookupEnv("SITE_URL")
	if !present {
		log.Panic("SITE_URL must be defined in environment")
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index)

	n := negroni.Classic() // Includes some default middlewares
	n.UseHandler(mux)

	go updateSiteStatus()

	n.Run(":" + port)
}
