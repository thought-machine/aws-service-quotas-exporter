subinclude("///third_party/subrepos/pleasings//docker")

docker_image(
    name = "aws-quotas-exporter_alpine",
    srcs = [
        "//cmd:aws-service-quotas-exporter",
    ],
    dockerfile = "Dockerfile-aws-service-quotas-exporter",
    image = "aws-service-quotas-exporter",
    visibility = [
        "//k8s",
    ],
)
