FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY querypie-mcp-server /usr/bin/querypie-mcp-server
ENTRYPOINT ["/usr/bin/querypie-mcp-server"]
