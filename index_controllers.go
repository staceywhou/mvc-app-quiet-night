package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"strings"
	"github.com/go-chi/chi"
)

func indexController(w http.ResponseWriter, r *http.Request) {

	referer := r.Header.Get("Referer")

	incrementCount(referer)

	type indexContextData struct {
		Events []Event
		Today  time.Time
	}

	theEvents, err := getAllEvents()
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	contextData := indexContextData{
		Events: theEvents,
		Today:  time.Now(),
	}

	tmpl["index"].Execute(w, contextData)
}

func aboutController(w http.ResponseWriter, r *http.Request) {
	type aboutContextData struct {
		// Add any data needed for the "about" page
	}

	tmpl["about"].Execute(w, aboutContextData{})
}

// isYaleEmail checks if the email address ends with "yale.edu"
func isYaleEmail(email string) bool {
	// Use a simple regular expression for demonstration
	// You might want to use a more comprehensive email validation library
	match, _ := regexp.MatchString(`@yale\.edu$`, email)
	return match
}

func detailsController(w http.ResponseWriter, r *http.Request) {

	// if method is get, show the events:
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid event ID", http.StatusBadRequest)
		return
	}

	event, found := getEventByID(eventID)
	if !found {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	type detailsContextData struct {
		Event  Event
		Today  time.Time
		Errors string
	}

	contextData := detailsContextData{
		Event: event,
		Today: time.Now(),
	}

	// allow people to RSVP
	if r.Method == http.MethodPost {

		// Parse form data
		r.ParseForm()

		// Validate email format
		email := r.FormValue("email")
		if !isYaleEmail(email) {
			contextData.Errors = "must be a Yale email address"
			tmpl["details"].Execute(w, contextData)
			return
		}

		// Get the event ID from the URL
		eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid event ID", http.StatusBadRequest)
			return
		}

		// add new attendee to the event
		addAttendee(eventID, r.FormValue("email"))

		// Redirect to the event details page
		http.Redirect(w, r, "/events/"+strconv.Itoa(eventID), http.StatusFound)
		return
	}
	// Execute the HTML template
	tmpl["details"].Execute(w, contextData)

}

func incrementCount(domain string) {
	// Increment the count in the database
	_, err := db.Exec(`
		INSERT INTO visitors (domain, count)
		VALUES (?, 1)
		ON CONFLICT(domain) DO UPDATE SET count = count + 1
	`, domain)
	if err != nil {
		log.Println(err)
	}
}
func getAllVisitors() ([]Visitor, error) {
	rows, err := db.Query("SELECT domain, count FROM visitors ORDER BY count DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visitors []Visitor

	for rows.Next() {
		var v Visitor
		err := rows.Scan(&v.Domain, &v.Count)
		if err != nil {
			return nil, err
		}
		visitors = append(visitors, v)
	}

	return visitors, nil
}

type Visitor struct {
	Domain string
	Count  int
}

func visitorController(w http.ResponseWriter, r *http.Request) {

	visitors, err := getAllVisitors()

	if err != nil {
		log.Println(err)
	}

	data := map[string]interface{}{
		"Visitors": visitors,
	}

	// Execute the HTML template
	tmpl["visitor"].Execute(w, data)

}

func createController(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse form data
		r.ParseForm()

		// Validate title
		title := r.FormValue("title")
		if len(title) < 5 || len(title) > 50 {
			http.Error(w, "Title must be between 5 and 50 characters", http.StatusBadRequest)
			renderCreateTemplate(w, r, "error", "Title must be between 5 and 50 characters")
			return
		}

		// Validate location
		location := r.FormValue("location")
		if len(location) < 5 || len(location) > 50 {
			http.Error(w, "Location must be between 5 and 50 characters", http.StatusBadRequest)
			renderCreateTemplate(w, r, "error", "Location must be between 5 and 50 characters")
			return
		}

		// Validate image
		image := r.FormValue("image")
		if !isValidImageURL(image) {
			http.Error(w, "Invalid image URL or format", http.StatusBadRequest)
			renderCreateTemplate(w, r, "error", "Invalid image URL or format")
			return
		}

		// Validate date
		dateStr := r.FormValue("date")
		date, err := time.Parse("2006-01-02T15:04", dateStr)
		if err != nil || date.Before(time.Now()) {
			http.Error(w, "Invalid date format or date is in the past", http.StatusBadRequest)
			renderCreateTemplate(w, r, "error", "Invalid date format or date is in the past")
			return
		}

		// Create an Event object from form data
		newEvent := Event{
			Title:    title,
			Location: location,
			Image:    image,
			Date:     date,
		}

		// Add the new event to the list
		addEvent(newEvent)

		// Convert newEvent.ID to string before concatenating with "/events/"
		eventIDStr := strconv.Itoa(newEvent.ID)

		// Redirect to the event detail page
		http.Redirect(w, r, "/events/"+eventIDStr, http.StatusFound)
		return
	}

	// Render the HTML template if HTTP GET
	renderCreateTemplate(w, r, "", "")
}

func renderCreateTemplate(w http.ResponseWriter, r *http.Request, class string, message string) {
	data := struct {
		Class   string
		Message string
	}{
		Class:   class,
		Message: message,
	}
	tmpl["create"].Execute(w, data)
}

func isValidImageURL(url string) bool {
	// Add your image URL validation logic here
	// For simplicity, you can check if the URL ends with one of the specified file extensions
	// You may want to use a more comprehensive approach in a production environment
	validExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".gifv"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(url, ext) {
			return true
		}
	}
	return false
}


// API Handler

func apiController(w http.ResponseWriter, r *http.Request) {
	// Get all events
	events, err := getAllEvents()
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Marshal events to JSON
	response := map[string][]Event{"events": events}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "error encoding JSON", http.StatusInternalServerError)
		return
	}

	// Set content type and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func apiEventController(w http.ResponseWriter, r *http.Request) {
	// Get the event ID from the URL
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid event ID", http.StatusBadRequest)
		return
	}

	// Get the specific event by ID
	event, found := getEventByID(eventID)
	if !found {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	// Marshal event to JSON
	jsonResponse, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "error encoding JSON", http.StatusInternalServerError)
		return
	}

	// Set content type and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}