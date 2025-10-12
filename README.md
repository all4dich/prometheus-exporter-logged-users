# prometheus-exporter-logged-users
## Prerequisites
* Install Go
  * [Download and install](https://go.dev/doc/install) 
* Install Bazel command
  * [gazelle](https://github.com/bazelbuild/bazel-gazelle)
  * [Bazel](https://bazel.build/install)

## Build 
* How to build
```shell
./build.sh
```
* build.sh
```shell
#!/bin/bash
bazel clean
bazel build //:deploy --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64
cp bazel-bin/prometheus-exporter-logged-users ./prometheus-exporter-logged-users
```
## Run
```shell
./prometheus-exporter-logged-users --port 19080 --token <your-token> --url http://localhost:8086 --org <your-org> --bucket <your-bucket>
```
## Help
```shell
./prometheus-exporter-logged-users --help
usage: prometheus-exporter-logged-users [-h|--help] [-p|--port <integer>]
                                        [-t|--token <string>]
                                        [-u|--url <string>] [-o|--org <string>]
                                        [-b|--bucket <string>]

                                        A Prometheus exporter for logged-in
                                        users

Arguments:

  -h  --help    Print help information
  -p  --port    Port number to start the server on. Default: 8080
  -t  --token   InfluxDB token
  -u  --url     InfluxDB URL
  -o  --org     InfluxDB Organization
  -b  --bucket  InfluxDB Bucket
```

## Metrics
The exporter exposes the following metrics:

* `logged_in_users`: The total number of users currently logged in.
* `logged_in_user`: Information about each logged-in user.
* `process_read_in_KB`: The read I/O of each process in KB/s.
* `process_write_in_KB`: The write I/O of each process in KB/s.
* `process_cpu_percent`: The CPU usage of each process as a percentage.
* `process_vsz`: The virtual memory size of each process.
* `process_rss`: The resident set size of each process.
