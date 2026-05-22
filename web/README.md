# web

Workspace contract for future web and dashboard work.

Why this directory exists now:

- preserve planned topology from the Charter architecture
- reserve a stable workspace ID so root task contracts remain fixed

Activation requirements:

- linked ADR for boundary changes if web introduces new shared contracts
- linked RFC before adding cross-cutting dashboard or hosted-audit behavior
- first implementation files should be `README.md`, `moon.yml`, app entrypoint, and verification tasks
- all checks must remain wrapped by the root Moon command family
