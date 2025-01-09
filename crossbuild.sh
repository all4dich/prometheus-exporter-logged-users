#env GOOS=darwin GOARCH=arm64 go build -o prometheus-exporter-logged-users-darwin-arm64 main.go
env GOOS=linux GOARCH=arm64 go build -o prometheus-exporter-logged-users-linux-arm64 main.go
#env GOOS=linux GOARCH=amd64 go build -o prometheus-exporter-logged-users-linux-amd64 main.go
