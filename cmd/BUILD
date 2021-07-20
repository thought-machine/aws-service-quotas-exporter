go_binary(
    name="aws-service-quotas-exporter",
    srcs=["main.go"],
    static=False,
    deps=[
        "//pkg/service_exporter:serviceexporter",
        "//third_party/go:prometheus",
        "//third_party/go:logrus",
        "//third_party/go:go-flags",
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
