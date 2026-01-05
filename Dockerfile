FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o leon-speedcams-go ./cmd

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /app/leon-speedcams-go .

CMD ["./leon-speedcams-go"]