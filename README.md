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
