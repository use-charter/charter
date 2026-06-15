# api/openapi/

Reserved for OpenAPI contracts, should Charter ever expose a public HTTP API.
**Empty today** — Charter is a CLI. Its machine-readable contracts are the JSON
Schemas in [`schemas/`](../../schemas/) (config + `doctor` result) and the SARIF
output, not an HTTP surface.

If an HTTP API is ever added, the spec lands here first (contract-first),
targeting OpenAPI 3.1+, with related ADR/RFC links and paired examples and tests.
