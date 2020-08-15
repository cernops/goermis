FROM golang:1.13.14-alpine3.12 

WORKDIR /go/src/gitlab.cern.ch/lb-experts/goermis/
RUN apk update && apk add --no-cache git \
    &&  mkdir -p  /var/lib/ermis/ /var/log/ermis/ 
 
COPY .  .
RUN go mod download
COPY staticfiles      /var/lib/ermis/staticfiles
COPY templates        /var/lib/ermis/templates

RUN go build -o /go/bin/goermis .
 
EXPOSE 8080
ENTRYPOINT ["/go/bin/goermis"]


