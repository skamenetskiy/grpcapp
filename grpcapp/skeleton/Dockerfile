FROM golang:1.19-alpine AS cache
WORKDIR /app
ADD go.mod .
ADD go.sum .
RUN go mod download

FROM cache AS build
RUN go -o bin/app.cmd build cmd/main.go

FROM alpine
COPY --from=build /app/bin/app.cmd /bin/app.cmd
CMD ["/bin/app.cmd"]