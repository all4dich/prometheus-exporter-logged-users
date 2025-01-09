#!/bin/bash
echo "Building prometheus-exporter-logged-users"
rm -vf prometheus-exporter-logged-users-linux-amd64 prometheus-exporter-logged-users-linux-arm64 prometheus-exporter-logged-users-darwin-arm64
echo "Building prometheus-exporter-logged-users-darwin-arm64"
env GOOS=darwin GOARCH=arm64 go build -o prometheus-exporter-logged-users-darwin-arm64 main.go
echo "Building prometheus-exporter-logged-users-linux-arm64"
env GOOS=linux GOARCH=arm64 go build -o prometheus-exporter-logged-users-linux-arm64 main.go
echo "Building prometheus-exporter-logged-users-linux-amd64"
env GOOS=linux GOARCH=amd64 go build -o prometheus-exporter-logged-users-linux-amd64 main.go
echo "Copying prometheus-exporter-logged-users for $(go env GOOS)-$(go env GOARCH)"
cp -v prometheus-exporter-logged-users-$(go env GOOS)-$(go env GOARCH) prometheus-exporter-logged-users