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

type Place struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type Highlight struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type Video struct {
	URL     string `json:"url"`
	Poster  string `json:"poster"`
	Caption string `json:"caption"`
	Credit  string `json:"credit,omitempty"` // Added this line
}

type Day struct {
	DayNumber   int    `json:"dayNumber"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Video       *Video `json:"video,omitempty"`
}

type Itinerary struct {
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	DurationDays  int         `json:"duration_days"`
	Hero          Hero        `json:"hero"`
	Highlights    []Highlight `json:"highlights"`
	PlacesVisited []Place     `json:"placesVisited"`
	DayByDay      []Day       `json:"dayByDay"`
	Inclusions    []string    `json:"inclusions"`
	Exclusions    []string    `json:"exclusions"`
	Specialist    Specialist  `json:"specialist"`
	SEO           SEO         `json:"seo"`
}

type Hero struct {
	Video  string `json:"video"`
	Poster string `json:"poster"`
}

type Specialist struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Photo    string `json:"photo"`
	Whatsapp string `json:"whatsapp"`
}

type SEO struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
}

var tmpl *template.Template

func main() {
	var err error
	tmpl, err = template.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Fatalf("FATAL: Error parsing templates: %v", err)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/itineraries/", itineraryDetailHandler)

	log.Println("Server starting on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("FATAL: Error starting server: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	err := tmpl.ExecuteTemplate(w, "home.tmpl", nil)
	if err != nil {
		log.Printf("ERROR: Failed to execute home template: %v", err)
		http.Error(w, "Server error: could not render homepage.", http.StatusInternalServerError)
	}
}

func itineraryDetailHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/itineraries/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join("content", "itineraries", slug+".json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("ERROR: Could not read file for slug '%s' at path %s: %v", slug, filePath, err)
		http.Error(w, "Server error: could not read itinerary file.", http.StatusNotFound)
		return
	}

	var itinerary Itinerary
	err = json.Unmarshal(file, &itinerary)
	if err != nil {
		log.Printf("CRITICAL: Failed to parse JSON for slug '%s'. ERROR: %v", slug, err)
		http.Error(w, "Server error: could not process itinerary data.", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "itinerary_detail.tmpl", itinerary)
	if err != nil {
		log.Printf("ERROR: Failed to execute template for slug '%s': %v", slug, err)
		http.Error(w, "Server error: could not render page.", http.StatusInternalServerError)
	}
}