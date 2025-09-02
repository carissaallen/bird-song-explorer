FROM golang:1.21-alpine AS builder

# Install certificates and git
RUN apk --no-cache add ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bird-song-explorer cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates ffmpeg
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bird-song-explorer .

COPY final_intros ./final_intros/
COPY final_outros ./final_outros/
COPY assets ./assets/
COPY sound_effects ./sound_effects/

EXPOSE 8080

# Run the binary
CMD ["./bird-song-explorer"]