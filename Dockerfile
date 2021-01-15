FROM golang:1.15-alpine AS build

RUN apk --no-cache add ca-certificates
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 go build -o /bin/service

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/service /bin/service
ENTRYPOINT ["/bin/service"]