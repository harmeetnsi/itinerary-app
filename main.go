import (
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type Itinerary struct {
    Title        string   `json:"title"`
    Slug         string   `json:"slug"`
    DurationDays int      `json:"duration_days"`
    Hero         struct {
        Video  string `json:"video"`
        Poster string `json:"poster"`
    } `json:"hero"`
    Highlights []string `json:"highlights"`
    Days       []struct {
        Day     int    `json:"day"`
        Title   string `json:"title"`
        Summary string `json:"summary"`
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
    // Serve static files
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Home page route
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    const tplPath = "/opt/itinerary-app/templates/home.tmpl"
    log.Println("Rendering template:", tplPath)
    tmpl, err := template.ParseFiles(tplPath)
    if err != nil {
        log.Println("template parse error:", err)
        http.Error(w, "template error", 500)
        return
    }
    if err := tmpl.Execute(w, nil); err != nil {
        log.Println("template execute error:", err)
    }
}

	func itinerariesHandler(w http.ResponseWriter, r *http.Request) {
    if !strings.HasPrefix(r.URL.Path, "/itineraries/") {
        http.NotFound(w, r)
        return
    }
    slug := strings.TrimPrefix(r.URL.Path, "/itineraries/")
    slug = strings.TrimSuffix(slug, "/")
    if slug == "" {
        http.NotFound(w, r)
        return
    }

    jsonPath := filepath.Join("/opt/itinerary-app/content/itineraries", slug+".json")
    data, err := os.ReadFile(jsonPath)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    var itin Itinerary
    if err := json.Unmarshal(data, &itin); err != nil {
        http.Error(w, "invalid itinerary file", http.StatusInternalServerError)
        return
    }

    tplPath := "/opt/itinerary-app/templates/itinerary_detail.tmpl"
    tpl := template.Must(template.ParseFiles(tplPath))
    if err := tpl.Execute(w, itin); err != nil {
        log.Println("template execute error:", err)
    }
})
// Itineraries route
http.HandleFunc("/itineraries/", itinerariesHandler)

log.Println("Listening on 127.0.0.1:8080")
log.Fatal(http.ListenAndServe(":8080", nil))

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}