FROM golang:1.13.14-alpine3.12 AS builder
WORKDIR /go/src/gitlab.cern.ch/lb-experts/goermis/
RUN apk update && apk add --no-cache git  
COPY .  .
RUN go mod download
RUN go build -o /go/bin/goermis .


FROM alpine:latest
LABEL maintainer="Kristian Kouros <kristian.kouros@cern.ch>"
WORKDIR /root/
RUN  mkdir -p  /var/lib/ermis/ /var/log/ermis/
COPY --from=builder   /go/bin/goermis   /
COPY staticfiles      /var/lib/ermis/staticfiles
COPY templates        /var/lib/ermis/templates
EXPOSE 8080
ENTRYPOINT ["/goermis"]


