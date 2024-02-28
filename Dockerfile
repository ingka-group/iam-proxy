FROM golang:1.22 as build

WORKDIR /go/src/

COPY . .

RUN make compile

FROM gcr.io/distroless/static

COPY --from=build /go/src/bin/iam-proxy /

# This would be nicer as `nobody:nobody` but distroless has no such entries.
USER 65535:65535
ENV HOME /

ENTRYPOINT ["/iam-proxy"]
