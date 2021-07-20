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
