package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Itinerary holds all the data for a single tour itinerary.
type Itinerary struct {
	Title         string        `json:"title"`
	Slug          string        `json:"slug"`
	DurationDays  int           `json:"duration_days"`
	Hero          Hero          `json:"hero"`
	SEO           SEO           `json:"seo"`
	Highlights    []Highlight   `json:"highlights"`
	PlacesVisited []Place       `json:"placesVisited"`
	DayByDay      []Day         `json:"dayByDay"`
	Inclusions    []string      `json:"inclusions"`
	Exclusions    []string      `json:"exclusions"`
	Specialist    Specialist    `json:"specialist"`
	OtherTours    []TourCard    `json:"otherTours"` // <-- ADDED THIS LINE
}

// Hero contains the video and poster image for the hero section.
type Hero struct {
	Video  string `json:"video"`
	Poster string `json:"poster"`
}

// SEO contains the meta title and description for the page.
type SEO struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
}

// Highlight represents a single card in the "Trip Highlights" section.
type Highlight struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Icon        string `json:"icon"`
}

// Place represents a location visited on the tour.
type Place struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// Day represents a single day in the itinerary, which can be text or video.
type Day struct {
	DayNumber   int    `json:"dayNumber,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Video       *Video `json:"video,omitempty"`
}

// Video represents an embedded video for a specific day.
type Video struct {
	URL     string `json:"url"`
	Poster  string `json:"poster"`
	Caption string `json:"caption"`
	Credit  string `json:"credit,omitempty"`
}

// Specialist contains information about the tour specialist.
type Specialist struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Photo    string `json:"photo"`
	Whatsapp string `json:"whatsapp"`
}

// TourCard represents a card in the "Explore Other Tours" section.
type TourCard struct {
	Title string `json:"title"`
	Image string `json:"image"`
	URL   string `json:"url"`
}

var templates *template.Template

func main() {
	var err error
	templates, err = template.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Fatalf("Could not parse templates: %v", err)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/itineraries/", itineraryHandler)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.tmpl", nil)
	if err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Server error: could not render page.", http.StatusInternalServerError)
	}
}

func itineraryHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/itineraries/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join("content", "itineraries", slug+".json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Could not read itinerary file for slug '%s': %v", slug, err)
		http.NotFound(w, r)
		return
	}

	var itinerary Itinerary
	if err := json.Unmarshal(file, &itinerary); err != nil {
		log.Printf("Could not parse itinerary JSON for slug '%s': %v", slug, err)
		http.Error(w, "Server error: invalid itinerary file.", http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "itinerary_detail.tmpl", itinerary)
	if err != nil {
		log.Printf("Error executing itinerary template for slug '%s': %v", slug, err)
		http.Error(w, "Server error: could not render page.", http.StatusInternalServerError)
	}
}