go_binary(
    name = "aws-service-quotas-exporter",
    srcs = ["cmd/main.go"],
    static = False,
    deps = [
        "//pkg/service_exporter:serviceexporter",
        "//third_party/go:prometheus",
        "//third_party/go:logrus",
        "//third_party/go:go-flags",
    ],
)
