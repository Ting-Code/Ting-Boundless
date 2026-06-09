# syntax=docker/dockerfile:1
# Generic multi-stage build for any Go service in this monorepo.
# Usage: docker build -f deploy/Dockerfile --build-arg SERVICE=gateway .

FROM golang:1.25-alpine AS build
WORKDIR /src/go
COPY go/go.mod go/go.sum ./
RUN go mod download
COPY go/ .
ARG SERVICE
RUN test -n "$SERVICE" && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/app ./services/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
