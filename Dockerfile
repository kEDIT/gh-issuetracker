# app build stage

FROM golang:1.20.8-alpine3.18 as Builder

WORKDIR /build

COPY . . 

RUN go mod download

EXPOSE 8090

RUN CGO_ENABLED=0 go build -o main main.go

ENTRYPOINT [ "./main" ]
