# cloud/

Reserved for a future hosted control plane — a managed layer that would sit on
top of the open-source CLI. **Nothing is built here yet**; the directory holds
only its Moon workspace config so the build graph stays stable.

Charter is, and will remain, a fully offline CLI. Anything added here would be
**additive and optional**, and would land behind:

- an ADR for its trust boundaries and data ownership, and
- an RFC before any network service, persistence, or hosted governance flow.

Until then there is nothing to run or configure in this directory.
