package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var allEvents []Event

// A "mutex" is used to say "hey, I'm using these data,
// don't touch them while I'm using them!" We use the
// mutex when adding events or attendees.
var allEventsMutex = &sync.Mutex{}

// Event - encapsulates information about an event
type Event struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Location  string     `json:"location"`
	Image     string     `json:"image"`
	Date      time.Time  `json:"date"`
	Attending []Attendee `json:"attending"`
}

// Attendee struct representing a person attending an event
type Attendee struct {
	Email            string `json:"email"`
	ConfirmationCode string `json:"confirmationCode"`
}

// getEventByID - returns the event in `allEvents` that has
// the specified id and a boolean indicating whether or not
// it was found. If it is not found, returns an empty event
// and false.
func getEventByID(id int) (Event, bool) {
	for _, event := range allEvents {
		if event.ID == id {
			return event, true
		}
	}
	return Event{}, false
}

// getAllEvents - returns slice of all events and an error status. Here it is
// just returns `nil` always for the error. In mgt660, we're using similar
// code that might actually return an error, but here it's always `nil`.
func getAllEvents() ([]Event, error) {
	return allEvents, nil
}

// getMaxEventID returns the maximum of all
// the ids of the events in allEvents
func getMaxEventID() int {
	maxID := -1
	for _, event := range allEvents {
		if event.ID > maxID {
			maxID = event.ID
		}
	}
	return maxID
}

// Adds an attendee to an event
func addAttendee(id int, email string) error {
	// When get a "lock", we are saying that we're
	// going to edit some data and we don't want
	// anybody else to use those data while we're
	// writing it. Recall, in go there might be
	// multiple threads (goroutines).
	allEventsMutex.Lock()
	defer allEventsMutex.Unlock()
	for i := 0; i < len(allEvents); i++ {
		if allEvents[i].ID == id {
			confirmationCode, err := generateConfirmationCode(email)
			if err != nil {
				return err
			}

			newAttendee := Attendee{
				Email:            email,
				ConfirmationCode: confirmationCode,
			}

			allEvents[i].Attending = append(allEvents[i].Attending, newAttendee)
			return nil
		}
	}
	return errors.New("no such event")
}

// Generates a confirmation code from the first 7 characters of the hash of the email
func generateConfirmationCode(email string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(email))
	if err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	// Take the first 7 characters as the confirmation code
	confirmationCode := hashString[:7]
	return confirmationCode, nil
}

// Add an event to the list of events.
// This uses a "mutex" to say "hey, I'm using these data,
// don't touch them while I'm using them!"
func addEvent(event Event) {
	allEventsMutex.Lock()
	event.ID = getMaxEventID() + 1
	allEvents = append(allEvents, event)
	allEventsMutex.Unlock()
}

// init is run once when this file is first loaded. See
// https://golang.org/doc/effective_go.html#init
// https://medium.com/golangspec/init-functions-in-go-eac191b3860a
func init() {
	newYorkTimeZone, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("Could not load timezone database on your system!")
	}

	defaultEvents := []Event{
		{
			ID:       1,
			Title:    "SOM House Party",
			Date:     time.Date(2023, 10, 17, 16, 30, 0, 0, newYorkTimeZone),
			Image:    "http://i.imgur.com/pXjrQ.gif",
			Location: "Kyle's house",
		},
		{
			ID:       2,
			Title:    "BBQ party for hackers and nerds",
			Date:     time.Date(2023, 10, 19, 19, 0, 0, 0, newYorkTimeZone),
			Image:    "http://i.imgur.com/7pe2k.gif",
			Location: "Judy Chevalier's house",
		},
		{
			ID:       3,
			Title:    "BBQ for managers",
			Date:     time.Date(2023, 12, 2, 18, 0, 0, 0, newYorkTimeZone),
			Image:    "http://i.imgur.com/CJLrRqh.gif",
			Location: "Barry Nalebuff's house",
		},
		// Here I didn't include an even #4 just to show that
		// events in a real system might be deleted and so you
		// would need to handle such cases. Eg. if somebody
		// tries to get event #4, you would typically return
		// a 404 error which means "not found".
		{
			ID:       5,
			Title:    "Cooking lessons for the busy business student",
			Date:     time.Date(2023, 12, 21, 19, 0, 0, 0, newYorkTimeZone),
			Image:    "http://i.imgur.com/02KT9.gif",
			Location: "Yale Farm",
		},
	}
	allEvents = append(allEvents, defaultEvents...)
	addAttendee(1, "kyle.jensen@yale.edu")
	addAttendee(1, "kim.kardashian@yale.edu")
	addAttendee(2, "kyle.jensen@yale.edu")
	addAttendee(2, "kim.kardashian@yale.edu")
	addAttendee(3, "kim.kardashian@yale.edu")
	addAttendee(5, "homer.simpson@yale.edu")
}
