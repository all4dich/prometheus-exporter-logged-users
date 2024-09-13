# prometheus-exporter-logged-users
## Build 
```shell
go get
go build
```
## Run
```shell
./prometheus-exporter-logged-users --port 19080
```
## Help
```shell
./prometheus-exporter-logged-users --help
usage: prometheus-exporter-logged-users [-h|--help] [-p|--port <integer>]

                                        A Prometheus exporter for logged-in
                                        users

Arguments:

  -h  --help  Print help information
  -p  --port  Port number to start the server on. Default: 8080
```
