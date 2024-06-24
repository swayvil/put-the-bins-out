package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const startYear = 2024
const nbOfYears = 10
const timeLocation = "Europe/Paris"
const calendarSummary = "Poubelles"
const eventSummary = "Sortir les poubelles de verre (4ème jeudi)"
const httpPort = "3000"

func main() {
	// Start an HTTP server listening to port 3000
	go startHTTPServer()

	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// CalendarScope allows: See, edit, share, and permanently delete all the calendars you can access using Google Calendar
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	createCalendarWithEvents(srv)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	// Get token from the URL: code=4/XXX
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func startHTTPServer() {
	http.HandleFunc("/", handleMain)
	fmt.Println(http.ListenAndServe(":"+httpPort, nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if len(code) > 0 {
		fmt.Println("Copy/Past this token and press enter: " + code)
	}
}

func addCalendar(srv *calendar.Service, summary string) string {
	// Set up the calendar details
	calendar := &calendar.Calendar{
		Summary:  summary,
		TimeZone: timeLocation,
	}

	// Use the calendar service client to create the calendar
	calendar, err := srv.Calendars.Insert(calendar).Do()
	if err != nil {
		log.Fatalf("Unable to create calendar: %v", err)
	}

	fmt.Printf("Calendar created: %s\n", calendar.Id)
	// Wait 2 seconds
	time.Sleep(2 * time.Second)
	return calendar.Id
}

func addEvent(srv *calendar.Service, calendarID string, summary string, dateTime string) {
	// Set up the event details
	event := &calendar.Event{
		Summary:     summary,
		Location:    "",
		Description: "",
		Start: &calendar.EventDateTime{
			DateTime: dateTime,
			TimeZone: timeLocation,
		},
		End: &calendar.EventDateTime{
			DateTime: dateTime,
			TimeZone: timeLocation,
		},
		Reminders: &calendar.EventReminders{
			UseDefault:      false,
			ForceSendFields: []string{"UseDefault"},
			Overrides: []*calendar.EventReminder{
				{Method: "popup", Minutes: 10},
			},
		},
	}

	// Use the calendar service client to create the event
	event, err := srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v", err)
	}
	fmt.Printf("Event created: %s\n", event.Start.DateTime)
}

func createCalendarWithEvents(srv *calendar.Service) {
	calendarId := addCalendar(srv, calendarSummary)

	// Add an event every fourth Wednesday of each month
	for year := startYear; year < startYear+nbOfYears; year++ {
		for month := 1; month <= 12; month++ {
			wednesday := searchDayBeforeFourthThursday(year, month)
			//fmt.Printf("Wednesday before 4th Thursday of %s: %s\n", wednesday.Month(), wednesday.Format(time.RFC3339))
			addEvent(srv, calendarId, eventSummary, wednesday.Format(time.RFC3339))
		}
		fmt.Printf("\n")
	}
}

func searchDayBeforeFourthThursday(year int, month int) time.Time {
	loc, err := time.LoadLocation(timeLocation)
	if err != nil {
		log.Fatalf("Unable to load time location: %v", err)
	}

	// Date on the first day of the month at 4pm
	startDate := time.Date(year, time.Month(month), 1, 16, 0, 0, 0, loc)

	// Iterate through each day of the month
	thursday := 0
	for current := startDate; current.Month() == time.Month(month); current = current.AddDate(0, 0, 1) {
		if current.Format("Monday") == "Thursday" {
			thursday++
		}
		if thursday == 4 {
			return current.AddDate(0, 0, -1) // Return the day (Wednesday) before
		}
	}
	return time.Time{}
}
