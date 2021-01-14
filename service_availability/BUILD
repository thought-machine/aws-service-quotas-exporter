go_library(
    name = "serviceavailability",
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
