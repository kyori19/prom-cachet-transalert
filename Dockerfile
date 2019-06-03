FROM golang:1.12.5-alpine3.9
RUN apk update \
  && apk add git
WORKDIR /go/src/github.com/kyori19/prom-cachet-transalert
COPY . .
RUN go get
RUN apk del git
RUN go install github.com/kyori19/prom-cachet-transalert
WORKDIR /
EXPOSE 9136
CMD ["prom-cachet-transalert"]