FROM golang:1.14 as builder

WORKDIR	/go/src/github.com/container-object-storage-interface/ephemeral-csi-driver

ADD ./bin/main /go/src/github.com/container-object-storage-interface/ephemeral-csi-driver/bin/main

ENTRYPOINT ["./bin/main"]
