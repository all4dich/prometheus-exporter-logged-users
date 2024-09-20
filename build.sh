#!/bin/bash
bazel clean
bazel build //:deploy --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64
cp bazel-bin/prometheus-exporter-logged-users ./prometheus-exporter-logged-users
bazel clean
bazel build //:deploy --platforms=@io_bazel_rules_go//go/toolchain:darwin_arm64
cp bazel-bin/prometheus-exporter-logged-users ./prometheus-exporter-logged-users-darwin-arm64
