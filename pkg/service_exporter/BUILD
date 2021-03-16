go_library(
    name = "serviceexporter",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    visibility = ["//..."],
    deps = [
        "//pkg/service_quotas:servicequotas",
        "//third_party/go:errors",
        "//third_party/go:logrus",
        "//third_party/go:prometheus",
    ]
)

go_test(
    name = "test",
    srcs = glob(["*_test.go"]),
    deps = [
        ":serviceexporter",
        "//third_party/go:testify",
    ],
)
