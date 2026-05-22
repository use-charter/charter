# action

Workspace contract for future GitHub Action packaging and distribution logic.

Why this directory exists now:

- preserve the end-state topology from the Charter architecture
- reserve a stable workspace ID for Moon task routing and CODEOWNERS

Activation requirements:

- linked ADR for packaging boundaries
- linked RFC if distribution model or runtime contract changes
- first implementation files should be `README.md`, `moon.yml`, action packaging source, and tests
- checks must route through the root Moon command family
