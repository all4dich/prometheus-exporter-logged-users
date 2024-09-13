package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

// Declare 'port' variable and assign it to 8080
// Get port nubmer from environment variable
var port = 18080

func getLoggedInUsers() (string, error) {
	// Execute the 'who' command to get logged in users
	out, err := exec.Command("who").Output()
	if err != nil {
		return "", err
	}

	// Convert the output to a string and trim any extra spaces
	users := strings.TrimSpace(string(out))
	return users, nil
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getLoggedInUsers()
	if err != nil {
		http.Error(w, "Error fetching logged-in users", http.StatusInternalServerError)
		return
	}

	// Prepare Prometheus format metrics
	metrics := "# HELP logged_in_users List of currently logged-in users.\n"
	metrics += "# TYPE logged_in_users gauge\n"

	// Count the number of users (each line in the 'who' output represents one user)
	userCount := 0
	if users != "" {
		userCount = len(strings.Split(users, "\n"))
	}

	// Prometheus gauge metric format
	metrics += fmt.Sprintf("logged_in_users %d\n", userCount)
	// Loop users and print them
	for _, user := range strings.Split(users, "\n") {
		if user != "" {
			user_info := strings.Fields(user)
			user_name := user_info[0]
			tty := user_info[1]
			when := user_info[2]
			from_location := user_info[4][1:]
			from_location = from_location[:len(from_location)-1]
			metrics += fmt.Sprintf("logged_in_user{user=\"%s\", tty=\"%s\", when=\"%s\", from=\"%s\"} 1\n",
				user_name, tty, when, from_location)
		}
	}

	// Write response in Prometheus format
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

func main() {
	// Set up HTTP server and route the '/metrics' path to the metricsHandler function
	http.HandleFunc("/metrics", metricsHandler)

	// Start the HTTP server on port 8080
	//fmt.Println("Starting server on : %d ...", port)
	fmt.Printf("Starting server on : %d ...", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
