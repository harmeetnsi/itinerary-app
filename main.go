
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
})

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}