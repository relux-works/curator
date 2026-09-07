# Pi MCP scope conflict requiring operator intent

The original DoD says all three curator run environments launch with MCP. The authoritative environments.md 1.1 section 7.8 (local curator-spec/protocol/environments.md:1274) explicitly declares Pi 0.84.2 has no MCP channel, no file and no fragment mcp section. Curator implementation internal/envregistry/envregistry.go:261 is MCP:nil, and launcher SPEC section 4.1 repeats this boundary. Native Pi upstream design only provides native launch plans and does not add MCP to Pi.

This is a product scope conflict, not an implementation fallback opportunity. The current protocol can deliver managed native Pi without MCP and MCP for Claude/Codex. Requiring MCP inside Pi adds a separately designed native extension/bridge and protocol/adapter scope; it cannot be truthfully implemented by inventing flags or writing an unconsumed MCP file.

Operator decision pending: qualify the current DoD so MCP applies to supported Claude/Codex channels (Pi stays native without MCP), or explicitly add Pi MCP design and delivery scope. Continue all unaffected launcher, native Pi, package and board work while the decision is pending. Never assume elapsed time approves either option.
