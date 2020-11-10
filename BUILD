subinclude("///third_party/subrepos/pleasings//docker")

go_binary(
    name = "aws-quotas-exporter",
    srcs = ["main.go"],
    static = False,
    deps = [
        "//third_party/go:prometheus",
        "//third_party/go:logrus",
        "//third_party/go:go-flags",
    ],
)

docker_image(
    name = "aws-quotas-exporter_alpine",
    srcs = [
        ":aws-quotas-exporter",
    ],
    dockerfile = "Dockerfile-aws-quotas-exporter",
    image = "aws-quotas-exporter",
    visibility = [
        "//k8s",
    ],
)
