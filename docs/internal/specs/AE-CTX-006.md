# AE-CTX-006

- Severity: Low (informational — does **not** deduct points or affect score)
- Category: Context
- Description: Agent instructions should be concise and declarative. Instruction-following research shows that stacked emphatic directives (`IMPORTANT`/`NEVER`/`MUST`/`CRITICAL`/…) create a fragile, competitive instruction topology that *degrades* adherence, while declarative phrasing transfers more reliably. AE-CTX-006 is a quality nudge for context files that over-use emphatic directives.
- Detection logic: reads the highest-precedence agent context file present (same resolution as AE-CTX-001). Counts word-boundary matches of `IMPORTANT|NEVER|MUST|CRITICAL|ALWAYS|EXTREMELY|ABSOLUTELY|FORBIDDEN|PROHIBITED` and divides by word count × 1,000. Flags when density ≥ 15 per 1,000 words.
- Pass example: a declarative AGENTS.md ("Tests run with `go test`. No silent mutation. Fail fast.") — low emphatic density, passes.
- Fail example: an AGENTS.md that stacks `IMPORTANT`/`NEVER`/`MUST`/`CRITICAL` throughout — density ≥ 15/1K, flagged informational; evidence reports the count and density.
- Evidence expectations: emphatic-directive count, word count, density, and the threshold.
- Edge cases: fires only when a context file exists (absence is AE-CTX-001's concern); empty file → no finding.
- Remediation: prefer concise, declarative guidance over stacked emphatic directives; state constraints plainly.
- Scoring impact: informational — listed in output, re-surfaces, but contributes **0** to the score (mirrors AE-SUPPRESS-003).
- Related ADRs: ADR-0023, ADR-0013
- Related evals: None yet
