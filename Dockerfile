# build stage
FROM golang:1.24-alpine AS build

# set working directory
WORKDIR /app

# copy source code
COPY go.mod go.sum ./

# install dependencies
RUN go mod download

# copy source code
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/leeta ./cmd/http/main.go

# final stage
FROM alpine:latest AS final
LABEL maintainer="emmrys-jay"

# set working directory
WORKDIR /app

# copy binary
COPY --from=build /app/bin/leeta ./

# copy config
COPY config.yml ./config.yml

# copy swagger folder for swagger ui   
COPY ./docs ./docs

EXPOSE 8080

ENTRYPOINT [ "./leeta" ]