load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "com_github_akamensky_argparse",
        importpath = "github.com/akamensky/argparse",
        sum = "h1:YGzvsTqCvbEZhL8zZu2AiA5nq805NZh75JNj4ajn1xc=",
        version = "v1.4.0",
    )
