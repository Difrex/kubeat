FROM golang AS builder

COPY . /go/src/github.com/Difrex/kubeat

WORKDIR /go/src/github.com/Difrex/kubeat

RUN go get -t -v ./...

ENV CGO_ENABLED=0

RUN go build -v -a -ldflags '-extldflags "-static"' -o /kubeat && strip /kubeat

FROM alpine AS final

RUN apk update && apk add ca-certificates curl

COPY --from=builder /kubeat /

USER nobody

ENTRYPOINT ["/kubeat"]
