FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o mcp-querypie .

FROM --platform=$BUILDPLATFORM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/mcp-querypie .

ENTRYPOINT ["/app/mcp-querypie"]
