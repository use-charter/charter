# ADR-0002 Single Root Go Module

- Status: Accepted
- Context: Early multi-module splits create churn before boundaries are proven.
- Decision: Use one root Go module, no `go.work`, no extra modules during bootstrap.
- Consequences: Package structure must absorb growth until real extraction pressure appears.
