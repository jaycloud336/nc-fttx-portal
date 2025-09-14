package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// Municipality represents a North Carolina municipality
type Municipality struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"` // "City" or "County"
	PermitExpiration string `json:"permit_expiration"`
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	TurnaroundDays   int    `json:"turnaround_days"`
	PermitFee        string `json:"permit_fee"`
	Requirements     string `json:"requirements"`
	GISLink          string `json:"gis_link"`
	PermitPortalLink string `json:"permit_portal_link"`
}

// Sample data - North Carolina municipalities for FTTX permitting
var municipalities = []Municipality{
	{
		ID:               1,
		Name:             "Raleigh",
		Type:             "City",
		PermitExpiration: "6 months",
		ContactEmail:     "permits@raleighnc.gov",
		ContactPhone:     "(919) 996-3000",
		TurnaroundDays:   14,
		PermitFee:        "$150.00",
		Requirements:     "Right-of-way permit required for all FTTX installations. Traffic control plan mandatory for major thoroughfares.",
		GISLink:          "https://maps.raleighnc.gov/iMAPS/",
		PermitPortalLink: "https://raleighnc.gov/permits-and-development",
	},
	{
		ID:               2,
		Name:             "Charlotte",
		Type:             "City",
		PermitExpiration: "6 months",
		ContactEmail:     "rowpermit@charlottenc.gov",
		ContactPhone:     "(704) 336-2891",
		TurnaroundDays:   21,
		PermitFee:        "$200.00",
		Requirements:     "Comprehensive utility coordination required. Environmental impact assessment for sensitive areas.",
		GISLink:          "https://maps.charlotte.gov/",
		PermitPortalLink: "https://charlottenc.gov/Transportation/Programs/Pages/Right-of-Way-Permitting.aspx",
	},
	{
		ID:               3,
		Name:             "Durham",
		Type:             "City",
		PermitExpiration: "3 months",
		ContactEmail:     "publicworks@durhamnc.gov",
		ContactPhone:     "(919) 560-4326",
		TurnaroundDays:   10,
		PermitFee:        "$125.00",
		Requirements:     "Standard ROW application with fiber route plans. Coordination with Duke Energy required.",
		GISLink:          "https://durhamnc.maps.arcgis.com/apps/webappviewer/index.html",
		PermitPortalLink: "https://durhamnc.gov/1329/Right-of-Way-Permits",
	},
	{
		ID:               4,
		Name:             "Wake County",
		Type:             "County",
		PermitExpiration: "12 months",
		ContactEmail:     "row@wakegov.com",
		ContactPhone:     "(919) 856-6100",
		TurnaroundDays:   18,
		PermitFee:        "$175.00",
		Requirements:     "County-wide coordination required. Special provisions for unincorporated areas.",
		GISLink:          "https://maps.wakegov.com/",
		PermitPortalLink: "https://www.wakegov.com/departments-government/public-works/right-way-permits",
	},
}

var templates *template.Template

func main() {
	// Load templates
	var err error
	templates, err = template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatal("Error loading templates:", err)
	}

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/api/municipalities", municipalitiesAPIHandler)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 NC FTTX Portal starting on port %s", port)
	log.Printf("🌐 Access at: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title           string
		Municipalities  []Municipality
		TotalCount      int
	}{
		Title:          "NC FTTX Permitting Portal",
		Municipalities: municipalities,
		TotalCount:     len(municipalities),
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func municipalitiesAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"municipalities": municipalities,
		"count":          len(municipalities),
		"timestamp":      time.Now().UTC(),
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "nc-fttx-portal",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Simple metrics for monitoring integration
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP nc_fttx_municipalities_total Total number of municipalities\n")
	fmt.Fprintf(w, "# TYPE nc_fttx_municipalities_total gauge\n")
	fmt.Fprintf(w, "nc_fttx_municipalities_total %d\n", len(municipalities))
	
	fmt.Fprintf(w, "# HELP nc_fttx_http_requests_total Total HTTP requests\n")
	fmt.Fprintf(w, "# TYPE nc_fttx_http_requests_total counter\n")
	fmt.Fprintf(w, "nc_fttx_http_requests_total{method=\"GET\",endpoint=\"/\"} 1\n")
}

