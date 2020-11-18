go_library(
    name = "servicequotas",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    visibility = ["//..."],
    deps = [
        "//third_party/go:aws-sdk-go",
        "//third_party/go:errors",
    ],
)

go_test(
    name = "test",
    srcs = glob(["*_test.go"]),
    deps = [
        ":service_quotas",
        "//third_party/go:testify",
    ],
)
