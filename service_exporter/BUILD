go_library(
    name = "serviceexporter",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    visibility = ["//..."],
    deps = [
        "//service_quotas:servicequotas",
        "//third_party/go:errors",
        "//third_party/go:logrus",
        "//third_party/go:prometheus",
    ]
)
