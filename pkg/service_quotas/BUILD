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
        "//third_party/go:logrus"
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

sh_cmd(
    name = "lint",
    cmd = "golint -set_exit_status $SRCS",
    srcs = glob(["*.go"]),
    labels = ["lint"]
)

sh_cmd(
        name = "gofmt",
        cmd = "[ -z \"$(gofmt -l $SRCS)\" ] && exit 0 || exit 1",
        srcs = glob(["*.go"]),
        labels = ["gofmt"],
)
