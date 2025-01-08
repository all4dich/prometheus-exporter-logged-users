package main

import (
	"bufio"
	"fmt"
	"github.com/akamensky/argparse"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var port int

func readCgroupInfo(pid int) (string, string, string, error) {
	cgroupFile := filepath.Join("/proc", fmt.Sprint(pid), "cgroup")
	file, err := os.Open(cgroupFile)
	hierarchyID := ""
	subsystem := ""
	cgroupPath := ""
	if err != nil {
		return hierarchyID, subsystem, cgroupPath, fmt.Errorf("failed to open cgroup file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Printf("Read Cgroup information for PID %d:\n", pid)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		fmt.Println(fields)
		if len(fields) >= 3 {
			// fields[0] is the hierarchy ID, fields[1] is the subsystem, fields[2] is the cgroup path
			hierarchyID = fields[0]
			subsystem = fields[1]
			cgroupPath = fields[2]
			if hierarchyID == "0" {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return hierarchyID, subsystem, cgroupPath, fmt.Errorf("error reading cgroup file: %w", err)
	}

	return hierarchyID, subsystem, cgroupPath, nil
}

func getContainerName(containerId string) (string, error) {
	containerName := ""
	//  docker inspect -f '{{.Name}}' <containerId>>?
	out, err := exec.Command("docker", "inspect", "-f", "'{{.Name}}'", containerId).Output()
	if err != nil {
		fmt.Println("Error:", err)
		return containerName, err
	}
	containerName = strings.Trim(string(out), "'\n")
	return containerName, nil
}

func checkCgroup(pid int) (string, string, string, string, string, error) {
	hierarchyId, subsystem, cgroupPath, err := readCgroupInfo(pid)
	containerId := ""
	containerName := ""
	if err != nil {
		fmt.Println("Error:", err)
		return hierarchyId, subsystem, cgroupPath, containerId, containerName, err
	} else {
		cgroupPathFields := strings.Split(cgroupPath, "/")
		processCgroup := cgroupPathFields[1]
		if processCgroup == "docker" {
			fmt.Println("This process is running in a Docker container")
			containerId = cgroupPathFields[2]
		} else {
			fmt.Println("This process is running in system or user cgroup")
			processInfo := cgroupPathFields[2]
			processInfoFields := strings.Split(processInfo, "-")
			if processInfoFields[0] == "docker" {
				fmt.Println("This process is running in a Docker container")
				containerId = strings.ReplaceAll(processInfoFields[1], ".scope", "")
			}
		}
		if containerId != "" {
			containerName, err = getContainerName(containerId)
		}
		return hierarchyId, subsystem, cgroupPath, containerId, containerName, err
	}
}

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
	//fmt.Println("Getting logged in users")
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
	//fmt.Println("Getting processes")
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

func getProcesses_with_mem_cpu() (string, error) {
	//fmt.Println("Getting processes with mem and cpu")
	out, err := exec.Command("ps", "-eo", "user:30,pid,pcpu,vsz,rss,cmd", "--sort=-rss", "--no-headers").Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	processes := strings.TrimSpace(string(out))
	return processes, nil
}

func metricsHandler_push() {
	users, err := getLoggedInUsers()
	host_name, err := getHostname()
	my_processes, err_get_process := getProcesses()
	process_with_mem_cpu, err_get_process_mem_cpu := getProcesses_with_mem_cpu()

	fmt.Println(my_processes)
	if err != nil {
		fmt.Println("Error fetching logged-in users:", err)
		return
	}
	if err_get_process != nil {
		fmt.Println("Error fetching process:", err_get_process)
		return
	}
	if err_get_process_mem_cpu != nil {
		fmt.Println("Error fetching process with mem and cpu:", err_get_process_mem_cpu)
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
				// Create code to catch exception
				process_info := strings.Fields(process)
				fmt.Println(process_info)
				process_id := process_info[0]
				user_name := process_info[2]
				read_Ks := process_info[3]
				write_Ks := process_info[5]

				hierarchyId := ""
				subsystem := ""
				cgroupPath := ""
				containerId := ""
				containerName := ""
				err := error(nil)

				process_id_str, _ := strconv.Atoi(process_id)
				if hierarchyId, subsystem, cgroupPath, containerId, containerName, err = checkCgroup(process_id_str); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Printf("Hierarchy ID: %s, Subsystem: %s, Cgroup Path: %s, Container ID: %s, Container Name: %s\n", hierarchyId, subsystem, cgroupPath, containerId, containerName)
				}
				if containerId == "" {
					containerId = "0 N/A"
					containerName = "0 N/A"
				}

				if process_info[7] == "?unavailable?" {
					process_command := process_info[8:]
					process_command_str := strings.Join(process_command, " ")
					metrics += fmt.Sprintf("process_read_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, containerId, containerName, process_command_str, read_Ks)
					metrics += fmt.Sprintf("process_write_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, containerId, containerName, process_command_str, write_Ks)
				} else {
					swapin_percent := process_info[7]
					io_percent := process_info[9]
					process_command := process_info[11:]
					process_command_str := strings.Join(process_command, " ")
					metrics += fmt.Sprintf("process_read_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", swapin=\"%s\", io=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, swapin_percent, io_percent, process_command_str, read_Ks)
					metrics += fmt.Sprintf("process_write_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", swapin=\"%s\", io=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, swapin_percent, io_percent, process_command_str, write_Ks)
				}
			}
		}
	}
	if process_with_mem_cpu != "" {
		//fmt.Println("Loop: Getting processes with mem and cpu - start")
		for _, process := range strings.Split(process_with_mem_cpu, "\n") {
			if process != "" {
				process_info := strings.Fields(process)
				username := process_info[0]
				process_id := process_info[1]
				cpu_percent := process_info[2]
				vsz := process_info[3]
				rss := process_info[4]
				process_command := process_info[5:]
				process_command_str := strings.Join(process_command, " ")
				hierarchyId := ""
				subsystem := ""
				cgroupPath := ""
				containerId := ""
				containerName := ""
				err := error(nil)
				// drop if process_command_str starts with [ or / or < or >
				if strings.HasPrefix(process_command_str, "[") || strings.HasPrefix(process_command_str, "/") || strings.HasPrefix(process_command_str, "<") || strings.HasPrefix(process_command_str, ">") {
					continue
				}
				process_id_str, _ := strconv.Atoi(process_id)
				if hierarchyId, subsystem, cgroupPath, containerId, containerName, err = checkCgroup(process_id_str); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Printf("Hierarchy ID: %s, Subsystem: %s, Cgroup Path: %s, Container ID: %s, Container Name: %s\n", hierarchyId, subsystem, cgroupPath, containerId, containerName)
				}
				if containerId == "" {
					containerId = "0 N/A"
					containerName = "0 N/A"
				}
				metrics += fmt.Sprintf("process_cpu_percent{hostname=\"%s\", username=\"%s\", process_id=\"%s\", cpu_percent=\"%s\", vsz=\"%s\", rss=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
					host_name, username, process_id, cpu_percent, vsz, rss, containerName, containerId, process_command_str, cpu_percent)
				metrics += fmt.Sprintf("process_vsz{hostname=\"%s\", username=\"%s\", process_id=\"%s\", cpu_percent=\"%s\", vsz=\"%s\", rss=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
					host_name, username, process_id, cpu_percent, vsz, rss, containerName, containerId, process_command_str, vsz)
				metrics += fmt.Sprintf("process_rss{hostname=\"%s\", username=\"%s\", process_id=\"%s\", cpu_percent=\"%s\", vsz=\"%s\", rss=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
					host_name, username, process_id, cpu_percent, vsz, rss, containerName, containerId, process_command_str, rss)
			}
		}
	}
	// Print metrics to standard output
	fmt.Println(metrics)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getLoggedInUsers()
	host_name, err := getHostname()
	my_processes, err_get_process := getProcesses()
	process_with_mem_cpu, err_get_process_mem_cpu := getProcesses_with_mem_cpu()

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
	if err_get_process_mem_cpu != nil {
		http.Error(w, "Error fetching process with mem and cpu", http.StatusInternalServerError)
		fmt.Println(err_get_process_mem_cpu)
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
				// Create code to catch exception
				process_info := strings.Fields(process)
				fmt.Println(process_info)
				process_id := process_info[0]
				user_name := process_info[2]
				read_Ks := process_info[3]
				write_Ks := process_info[5]

				hierarchyId := ""
				subsystem := ""
				cgroupPath := ""
				containerId := ""
				containerName := ""
				err := error(nil)

				process_id_str, _ := strconv.Atoi(process_id)
				if hierarchyId, subsystem, cgroupPath, containerId, containerName, err = checkCgroup(process_id_str); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Printf("Hierarchy ID: %s, Subsystem: %s, Cgroup Path: %s, Container ID: %s, Container Name: %s\n", hierarchyId, subsystem, cgroupPath, containerId, containerName)
				}
				if containerId == "" {
					containerId = "0 N/A"
					containerName = "0 N/A"
				}

				if process_info[7] == "?unavailable?" {
					process_command := process_info[8:]
					process_command_str := strings.Join(process_command, " ")
					metrics += fmt.Sprintf("process_read_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, containerId, containerName, process_command_str, read_Ks)
					metrics += fmt.Sprintf("process_write_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, containerId, containerName, process_command_str, write_Ks)
				} else {
					swapin_percent := process_info[7]
					io_percent := process_info[9]
					process_command := process_info[11:]
					process_command_str := strings.Join(process_command, " ")
					metrics += fmt.Sprintf("process_read_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", swapin=\"%s\", io=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, swapin_percent, io_percent, process_command_str, read_Ks)
					metrics += fmt.Sprintf("process_write_in_KB{hostname=\"%s\", process_id=\"%s\", username=\"%s\", read=\"%s\", write=\"%s\", swapin=\"%s\", io=\"%s\", command=\"%s\"} %s\n",
						host_name, process_id, user_name, read_Ks, write_Ks, swapin_percent, io_percent, process_command_str, write_Ks)
				}
			}
		}
	}
	if process_with_mem_cpu != "" {
		//fmt.Println("Loop: Getting processes with mem and cpu - start")
		for _, process := range strings.Split(process_with_mem_cpu, "\n") {
			if process != "" {
				process_info := strings.Fields(process)
				username := process_info[0]
				process_id := process_info[1]
				cpu_percent := process_info[2]
				vsz := process_info[3]
				rss := process_info[4]
				process_command := process_info[5:]
				process_command_str := strings.Join(process_command, " ")
				hierarchyId := ""
				subsystem := ""
				cgroupPath := ""
				containerId := ""
				containerName := ""
				err := error(nil)
				// drop if process_command_str starts with [ or / or < or >
				if strings.HasPrefix(process_command_str, "[") || strings.HasPrefix(process_command_str, "/") || strings.HasPrefix(process_command_str, "<") || strings.HasPrefix(process_command_str, ">") {
					continue
				}
				process_id_str, _ := strconv.Atoi(process_id)
				if hierarchyId, subsystem, cgroupPath, containerId, containerName, err = checkCgroup(process_id_str); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Printf("Hierarchy ID: %s, Subsystem: %s, Cgroup Path: %s, Container ID: %s, Container Name: %s\n", hierarchyId, subsystem, cgroupPath, containerId, containerName)
				}
				if containerId == "" {
					containerId = "0 N/A"
					containerName = "0 N/A"
				}
				metrics += fmt.Sprintf("process_cpu_percent{hostname=\"%s\", username=\"%s\", process_id=\"%s\", cpu_percent=\"%s\", vsz=\"%s\", rss=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
					host_name, username, process_id, cpu_percent, vsz, rss, containerName, containerId, process_command_str, cpu_percent)
				metrics += fmt.Sprintf("process_vsz{hostname=\"%s\", username=\"%s\", process_id=\"%s\", cpu_percent=\"%s\", vsz=\"%s\", rss=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
					host_name, username, process_id, cpu_percent, vsz, rss, containerName, containerId, process_command_str, vsz)
				metrics += fmt.Sprintf("process_rss{hostname=\"%s\", username=\"%s\", process_id=\"%s\", cpu_percent=\"%s\", vsz=\"%s\", rss=\"%s\", container_name=\"%s\", container_id=\"%s\", command=\"%s\"} %s\n",
					host_name, username, process_id, cpu_percent, vsz, rss, containerName, containerId, process_command_str, rss)
			}
		}
	}
	// Write response in Prometheus format
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

func main() {
	// Get a user id who call this program
	uid := os.Getuid()
	user, _ := user.LookupId(strconv.Itoa(uid))
	fmt.Println("User ID:", uid)
	fmt.Println("User Name:", user.Username)
	if uid != 0 {
		fmt.Println("You must run this program as root")
		os.Exit(1)
	}

	parser := argparse.NewParser("prometheus-exporter-logged-users", "A Prometheus exporter for logged-in users")
	portPtr := parser.Int("p", "port", &argparse.Options{Required: false, Help: "Port number to start the server on", Default: 8080})
	// Set up HTTP server and route the '/metrics' path to the metricsHandler function
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	port = *portPtr
	http.HandleFunc("/metrics", metricsHandler)

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			metricsHandler_push()
		}
	}()

	// Start the HTTP server on port $port
	fmt.Printf("Starting logged users collector server on : %d ...\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
