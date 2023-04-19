FROM golang:latest as build
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o bin/app .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/bin/app /app
EXPOSE 8000
CMD ["./app"]