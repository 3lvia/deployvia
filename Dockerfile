FROM golang:alpine AS build
LABEL maintainer="elvia@elvia.no"

ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./out/executable ./cmd/deployvia


FROM alpine:latest
LABEL maintainer="elvia@elvia.no"

RUN apk update && \
    apk upgrade --no-cache

# CVE-2024-12797
RUN apk add --no-cache libssl3 libcrypto3

RUN apk add openssh

RUN addgroup application-group --gid 1001 && \
    adduser application-user --uid 1001 \
        --ingroup application-group \
        --disabled-password

WORKDIR /app

COPY --from=build /app/out .

RUN chown --recursive application-user .
USER application-user

RUN mkdir -p ${HOME}/.ssh && \
    ssh-keyscan -H github.com >> ${HOME}/.ssh/known_hosts

EXPOSE 8080

ENTRYPOINT ["./executable"]
