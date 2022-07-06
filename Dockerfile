# build stage
FROM golang:1.18-alpine3.15 AS build-env
ADD . /src/github.com/ringtail/lucas
WORKDIR /src/github.com/ringtail/lucas
RUN go build -o app


# test stage
#FROM golang:1.8-alpine3.6
#WORKDIR /src/github.com/ringtail/lucas
#RUN go test


# release stage
FROM golang:1.18-alpine3.15
WORKDIR /bin
EXPOSE 8080
COPY --from=build-env /src/github.com/ringtail/lucas/app /bin/
CMD ["app","run"]
