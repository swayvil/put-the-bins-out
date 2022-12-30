# put-the-bins-out
Add a reminder in Google Calendar to put the bins out every fourth Wednesdays of each month.

# How to
1. On Google Cloud, [create a new project](https://console.cloud.google.com/projectcreate) and name it "put-the-bins-out"
2. In Google Cloud console, [enable the Google Calendar API for this project](https://console.cloud.google.com/apis/enableflow?apiid=calendar-json.googleapis.com&project=put-the-bins-out)
3. Authorize credentials for a desktop application. Follow the instructions [detailed in the quickstart](https://developers.google.com/calendar/api/quickstart/go)
4. On OAuth consent screen add the scope "./auth/calendar.app.created" and a test user
5. On Client ID for Desktop download the JSON file as credentials.json and move it to the project directory
6. Edit credentials.json and add the port ":3000"
```
"redirect_uris":["http://localhost:3000"]
```
7. Run the project:
```
go run put-the-bins-out.go
```

# References
Based on [golang quickstart](https://developers.google.com/calendar/api/quickstart/go)
