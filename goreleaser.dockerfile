FROM scratch
COPY querypie-mcp-server /usr/bin/querypie-mcp-server
ENTRYPOINT ["/usr/bin/querypie-mcp-server"]