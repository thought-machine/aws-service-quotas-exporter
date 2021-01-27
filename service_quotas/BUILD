go_library(
    name = "servicequotas",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go", "mock_ec2_client.go"],
    ),
    visibility = ["//..."],
    deps = [
        "//third_party/go:aws-sdk-go",
        "//third_party/go:errors",
    ],
)

go_test(
    name = "test",
    srcs = glob(["*_test.go", "mock_ec2_client.go"]),
    deps = [
        ":servicequotas",
        "//third_party/go:testify",
    ],
)
