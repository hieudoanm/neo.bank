FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12 AS runtime
COPY --from=build /out/app /app
COPY --from=build /out/migrate /migrate
COPY --from=build /src/internal/db/migrations /internal/db/migrations
USER nonroot:nonroot
ENTRYPOINT ["/app"]
