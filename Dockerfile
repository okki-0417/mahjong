FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mahjongd ./cmd/mahjongd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mahjongd /mahjongd
EXPOSE 8080
ENTRYPOINT ["/mahjongd"]
