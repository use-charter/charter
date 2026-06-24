---
title: "The harness is mostly your repo"
description: "Addy Osmani frames an agent as ~10% model and 90% harness. A large slice of that harness lives in your repository — and repo-readiness should be measured, not vibed."
date: 2026-06-24
author: "Charter"
tags: ["ai-agents", "engineering"]
---

Addy Osmani's [new-SDLC piece](https://addyosmani.com/blog/new-sdlc-vibe-coding/) has a line that reframes the agent debate: an agent is roughly 10% model and 90% harness, and *most agent failures are configuration failures.*

His metaphor: the model is the engine; the harness is the car, the road, and the traffic laws. Swap the engine under a stable harness and little changes. The harness is what decides how the agent behaves.

I'm writing as a reader of that piece, not a collaborator on it — his framing just put precise words to something I'd already been building toward with Charter.

Read that as a repo person and one thing jumps out. A large slice of that 90% doesn't live in the agent. It lives in your repository.

## The harness is mostly your repo

`AGENTS.md`, the one true test command, pinned tools, the MCP allowlist, the CI gate, the rules for what's unsafe to touch — that's the harness, checked into git. The model is portable. This layer isn't; it's yours to build.

So when an agent behaves, it's rarely just a strong model — the repo gave it somewhere safe to stand. When it flails, the model is usually fine. The repo handed it a weak harness.

## Most agent failures are boring

The dramatic failure is hallucination: an invented API, a package that doesn't exist. It happens. But in real repos the common failures are duller —

- no `AGENTS.md`, no stated conventions
- no obvious test command, no reliable local setup
- unpinned tools, or hook config that can run anything
- secrets sitting where an agent can read them
- CI that gives the agent no clean way to verify its work
- MCP servers that can shift under the workflow

None of those are intelligence problems. They're configuration — exactly Addy's point. A strong model still loses in a repo that hands it a weak harness.

## Verification is the line

Addy's split between vibe coding and engineering doesn't hinge on whether AI is involved. It hinges on verification — *the differentiator is not whether you use AI, it is how outputs get verified.* Set the bar at the eval, not the demo.

That cuts at repo-readiness too. Most teams vibe-check it: Claude Code struggles here, Cursor's great there. Useful, not reproducible. If readiness matters, it shouldn't be a feeling — it should be a number.

## Repo readiness should be a number

We already measure the harness's neighbors: lint, coverage, dependency and secret scans, CI health. Readiness for agents is the missing gauge — same idea as a linter: boring, reproducible, gateable.

And it shouldn't be an LLM judging whether your repo is ready for an LLM. A score you can't reproduce is a vibe, not a gate.

## Where Charter fits

[Charter](https://github.com/use-charter/charter) is an offline CLI that scores a repo `0–100` on agent-readiness — across context, secrets, MCP/tool safety, environment, CI, tests, and governance — then points at the exact gaps.

No model in the scoring path, no network: same repo, same score, every time. Gate it in CI and the number turns into a floor — every agent session starts safer than the last.

The longer case for treating the repo as part of the prompt lives in [The repo is part of the prompt](/blog/the-repo-is-part-of-the-prompt); the [Charter launch write-up](/blog/introducing-charter) covers where it started.

## The model will keep changing

You'll probably swap models within the year. The repo-side harness stays — the context, the tests, the CI, the boundaries, the conventions. AI doesn't remove that layer; it raises the cost of getting it wrong.

Addy's lifecycle isn't just better models. It's better systems around them. Repo-readiness is one of those systems — not the flashiest, maybe one of the most necessary.

`brew install use-charter/tap/charter && charter doctor` — the first number usually stings. What should a repo prove before an agent works in it? That's the rule set I want torn apart.
