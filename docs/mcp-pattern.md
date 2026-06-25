# MCP Pattern

Portfolio Manager exposes LLM-facing capabilities through a central MCP server with module-local tool registration.

## Registration Shape

- Each module that exposes MCP tools owns a local `mcp.go`.
- The module-local file defines tool names, schemas, handlers, and `RegisterMCPTools`.
- `internal/server/mcp.go` wires module registrations into the shared MCP server.
- Application startup passes the required services into `server.NewMCPServer` from `cmd/portfolio/main.go`.

Existing examples:

- `internal/portfolio/mcp.go`
- `internal/blotter/mcp.go`
- `pkg/mdata/mcp.go`
- `pkg/rdata/mcp.go`

This keeps the MCP surface scalable: adding tools for a module should usually mean adding or extending that module's `mcp.go`, then making only a small registration change in `internal/server/mcp.go`.

## Safety

- Mark read-only tools with MCP read-only annotations where available.
- Mark write/delete tools as destructive where appropriate.
- Write operations, including insert, update, and delete, must require explicit user confirmation before mutating data.
- The local pattern is to add a `confirm` argument and only proceed when it is exactly `yes`; otherwise return a prompt explaining what would be changed.

## Context Discipline

- Prefer compact list tools with filters, ids, and limits rather than returning whole datasets by default.
- Use typed JSON payloads and structured schemas so LLM clients do not need to infer command formats from prose.
- Keep handler responses JSON where the result may be consumed by another tool or follow-up analysis.

## MCP vs CLI

For LLM-facing operations, prefer MCP over the portfolio-manager CLI. MCP tools are discoverable, schema-described, and can carry safety annotations plus confirmation gates. The CLI remains useful for humans and scripts, but using it as the primary LLM interface adds command construction, shell quoting, and output parsing overhead.

Add CLI commands when they serve human or automation workflows independently of LLM clients; do not add CLI commands solely to expose a capability to MCP.
