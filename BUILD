subinclude("///third_party/subrepos/pleasings//docker")

go_binary(
    name = "aws-service-quotas-exporter",
    srcs = ["main.go"],
    static = False,
    deps = [
        "//service_exporter:serviceexporter",
        "//third_party/go:prometheus",
        "//third_party/go:logrus",
        "//third_party/go:go-flags",
    ],
)

docker_image(
    name = "aws-quotas-exporter_alpine",
    srcs = [
        ":aws-service-quotas-exporter",
    ],
    dockerfile = "Dockerfile-aws-service-quotas-exporter",
    image = "aws-service-quotas-exporter",
    visibility = [
        "//k8s",
    ],
)
