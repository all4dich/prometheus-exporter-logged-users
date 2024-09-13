package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var port int

func getLoggedInUsers() (string, error) {
	// Execute the 'w' command to get logged in users
	out, err := exec.Command("w").Output()
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
		userCount = len(strings.Split(users, "\n")) - 2
	}

	// Prometheus gauge metric format
	metrics += fmt.Sprintf("logged_in_users %d\n", userCount)
	// Loop users and print them
	// Remove the first two lines of the 'who' output as they are headers
	for _, user := range strings.Split(users, "\n")[2:] {
		if user != "" {
			user_info := strings.Fields(user)
			user_name := user_info[0]
			tty := user_info[1]
			from_location := user_info[2]
			when := user_info[3]
			idle_time := user_info[4]
			jcpu_time := user_info[5]
			pcpu_time := user_info[6]
			what_command := user_info[7]
			metrics += fmt.Sprintf("logged_in_user{user=\"%s\", tty=\"%s\", from=\"%s\", when=\"%s\", idle=\"%s\", jcpu=\"%s\", pcpu=\"%s\", what=\"%s\"} 1\n",
				user_name, tty, from_location, when, idle_time, jcpu_time, pcpu_time, what_command)
		}
	}

	// Write response in Prometheus format
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

func main() {
	parser := argparse.NewParser("print", "Prints provided string to stdout")
	portPtr := parser.Int("p", "port", &argparse.Options{Required: false, Help: "Port number to start the server on", Default: 8080})
	// Set up HTTP server and route the '/metrics' path to the metricsHandler function
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	port = *portPtr
	http.HandleFunc("/metrics", metricsHandler)

	// Start the HTTP server on port $port
	fmt.Printf("Starting logged users collector server on : %d ...", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
