FROM golang:alpine AS build

# Копируем исходный код в Docker-контейнер
WORKDIR /server
COPY . .

RUN go build -mod=vendor cmd/slow/main.go

# Копируем на чистый образ
FROM alpine

COPY --from=build /server/main /main

#CMD ['./main', '-port=:80', '-metrics=:8080']
CMD './main'