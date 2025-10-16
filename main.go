package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	BaseDir       string
	TemplatesDir  string
	StaticDir     string
	ContentDir    string
	ListenAddress string
}

func loadConfig() Config {
	base := os.Getenv("ITIN_BASE_DIR")
	if base == "" {
		cwd, _ := os.Getwd()
		base = cwd
	}
	return Config{
		BaseDir:       base,
		TemplatesDir:  filepath.Join(base, "templates"),
		StaticDir:     filepath.Join(base, "static"),
		ContentDir:    filepath.Join(base, "content", "itineraries"),
		ListenAddress: "127.0.0.1:8080",
	}
}

// Itinerary model
type Itinerary struct {
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	DurationDays int    `json:"duration_days"`
	Hero         struct {
		Video  string `json:"video"`
		Poster string `json:"poster"`
	} `json:"hero"`
	Highlights []string `json:"highlights"`
	Days       []struct {
		Day     int    `json:"day"`
		Title   string `json:"title"`
		Summary string `json:"summary"`
		Details string `json:"details"`
		Video   string `json:"video"`  // optional per-day video
		Poster  string `json:"poster"` // optional poster for MP4
	} `json:"days"`
	OtherTours []struct {
		Title        string `json:"title"`
		Slug         string `json:"slug"`
		Image        string `json:"image"`
		DurationDays int    `json:"duration_days"`
	} `json:"other_tours"`
	Specialist struct {
		Name     string `json:"name"`
		Role     string `json:"role"`
		Photo    string `json:"photo"`
		Whatsapp string `json:"whatsapp"`
	} `json:"specialist"`
	WhyUs []string `json:"why_us"`
	SEO   struct {
		MetaTitle       string `json:"meta_title"`
		MetaDescription string `json:"meta_description"`
	} `json:"seo"`
}

func main() {
	cfg := loadConfig()

	homeTplPath := filepath.Join(cfg.TemplatesDir, "home.tmpl")
	itinTplPath := filepath.Join(cfg.TemplatesDir, "itinerary_detail.tmpl")

	homeTpl, err := template.ParseFiles(homeTplPath)
	if err != nil {
		log.Fatalf("parse home template: %v", err)
	}

	// Add functions used in the itinerary template
	itinTpl, err := template.
		New("itinerary_detail.tmpl").
		Funcs(template.FuncMap{
			"contains":  strings.Contains,
			"hasSuffix": strings.HasSuffix,
			"replace":   strings.Replace,
		}).
		ParseFiles(itinTplPath)
	if err != nil {
		log.Printf("warning: could not parse itinerary_detail.tmpl (%v). Detail pages will 500.", err)
	}

	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir(cfg.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Home
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		setSecurityHeaders(w)
		if err := homeTpl.Execute(w, nil); err != nil {
			log.Printf("home execute error: %v", err)
			http.Error(w, "template render error", http.StatusInternalServerError)
		}
	})

	// Itinerary Detail
	mux.HandleFunc("/itineraries/", func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)

		slug := strings.TrimPrefix(r.URL.Path, "/itineraries/")
		slug = strings.Trim(slug, "/")
		if slug == "" {
			http.NotFound(w, r)
			return
		}

		jsonPath := filepath.Join(cfg.ContentDir, slug+".json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var itin Itinerary
		if err := json.Unmarshal(data, &itin); err != nil {
			log.Printf("invalid itinerary file %s: %v", jsonPath, err)
			http.Error(w, "invalid itinerary file", http.StatusInternalServerError)
			return
		}

		if itinTpl == nil {
			http.Error(w, "template not ready", http.StatusInternalServerError)
			return
		}
		if err := itinTpl.Execute(w, itin); err != nil {
			log.Printf("itinerary template execute error: %v", err)
			http.Error(w, "template render error", http.StatusInternalServerError)
			return
		}
	})

	handler := withLogging(mux)

	log.Printf("Listening on %s (base=%s)", cfg.ListenAddress, cfg.BaseDir)
	if err := http.ListenAndServe(cfg.ListenAddress, handler); err != nil {
		log.Fatal(err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// Allow YouTube/Vimeo iframes; allow self-hosted media and https streams
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; "+
			"img-src 'self' data: https:; "+
			"media-src 'self' https:; "+
			"style-src 'self' 'unsafe-inline'; "+
			"script-src 'self'; "+
			"frame-src 'self' https://www.youtube.com https://youtube.com https://player.vimeo.com;")
}