load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies2():
    go_repository(
        name = "com_github_akamensky_argparse",
        importpath = "github.com/akamensky/argparse",
        commit = "c010b5110f13a60a702a9b415159c58873508130"
    )