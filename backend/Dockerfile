# Compile stage
FROM golang:1.19 AS build-env

ADD . /backend
WORKDIR /backend

RUN go mod tidy
RUN go build -o /server

# Final stage
FROM debian:buster

EXPOSE 3000

WORKDIR /
COPY --from=build-env /server /

COPY index.html /
COPY assets /assets

RUN apt-get update; apt-get install ca-certificates -y

CMD ["/server"]