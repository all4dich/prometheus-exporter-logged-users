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

func getHostname() (string, error) {
	// Execute the 'hostname' command to get the hostname
	out, err := exec.Command("hostname").Output()
	if err != nil {
		return "", err
	}

	// Convert the output to a string and trim any extra spaces
	hostname := strings.TrimSpace(string(out))
	return hostname, nil
}

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

func getProcesses() (string, error) {
	// Execute the 'ps' command to get processes
	//out, err := exec.Command("/usr/sbin/iotop", "--only -k -b -n 1").Output()
	out, err := exec.Command("/usr/sbin/iotop", "--processes", "-qqq", "--only", "-k", "-b", "-n", "1").Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Convert the output to a string and trim any extra spaces
	processes := strings.TrimSpace(string(out))
	return processes, nil
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getLoggedInUsers()
	host_name, err := getHostname()
	my_processes, err_get_process := getProcesses()

	fmt.Println(my_processes)
	if err != nil {
		http.Error(w, "Error fetching logged-in users", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	if err_get_process != nil {
		http.Error(w, "Error fetching process", http.StatusInternalServerError)
		fmt.Println(err_get_process)
		return
	}

	// Prepare Prometheus format metrics
	metrics := "# HELP logged_in_users List of currently logged-in users.\n"
	metrics += "# TYPE logged_in_users gauge\n"
	metrics += "# HELP process_read_write_in_KB List of currently running processes with Read/Write KB/s.\n"
	metrics += "# TYPE process_read_write_in_KB gauge\n"

	// Count the number of users (each line in the 'who' output represents one user)
	userCount := 0
	if users != "" {
		userCount = len(strings.Split(users, "\n")) - 2
	}

	// Prometheus gauge metric format
	metrics += fmt.Sprintf("logged_in_users{hostname=\"%s\"} %d\n", host_name, userCount)
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
			what_command := user_info[7:]
			what_command_str := strings.Join(what_command, " ")
			metrics += fmt.Sprintf("logged_in_user{hostname=\"%s\", user=\"%s\", tty=\"%s\", from=\"%s\", when=\"%s\", idle=\"%s\", jcpu=\"%s\", pcpu=\"%s\", what=\"%s\"} 1\n",
				host_name, user_name, tty, from_location, when, idle_time, jcpu_time, pcpu_time, what_command_str)
		}
	}
	if my_processes != "" {
		for _, process := range strings.Split(my_processes, "\n") {
			if process != "" {
				process_info := strings.Fields(process)
				fmt.Println(process_info)
				process_id := process_info[0]
				user_name := process_info[2]
				read_Ks := process_info[3]
				write_Ks := process_info[5]
				swapin_percent := process_info[7]
				io_percent := process_info[9]
				process_command := process_info[11:]
				process_command_str := strings.Join(process_command, " ")
				metrics += fmt.Sprintf("process_read_write_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", swapin=\"%s\", io=\"%s\", command=\"%s\"} 1\n",
					host_name, process_id, user_name, read_Ks, write_Ks, swapin_percent, io_percent, process_command_str)
			}
		}
	}
	// Write response in Prometheus format
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

func main() {
	parser := argparse.NewParser("prometheus-exporter-logged-users", "A Prometheus exporter for logged-in users")
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
