---
title: "Introducing Charter: AI-agent readiness, scored"
description: "AI coding agents are only as good as the repo they work in. Charter grades any repository 0–100 on how safely an agent can operate in it — offline, deterministic, and with a concrete fix for every gap."
date: 2026-06-15
author: "Charter"
tags: ["launch", "ai-agents"]
---

An AI coding agent inherits everything about the repository it lands in. The good and the bad. Give it a clear `AGENTS.md`, pinned tools, and tests it can run, and it behaves like a careful senior engineer. Drop it into a repo with missing context, an unpinned MCP server, and a secret sitting in plain sight, and it will confidently make things worse — and you'll find out in the diff, not before.

That gap is invisible today. There's no number for "how ready is this repo for an agent." So we built one.

## What Charter measures

Charter is an offline CLI that scans a repository and returns a deterministic **0–100 readiness score** across nine categories — context, secrets, MCP safety, agent config, environment, CI, testing, autonomy, and governance. Eighteen rules, each with a severity that sets its weight, and each paired with the exact thing it checks and how to fix it.

It runs in under two seconds on a 50,000-file repo, makes **zero network calls**, and never sends your code anywhere. The same check runs in your shell and in CI.

## Why deterministic, not an LLM

It would be easy to point a model at a repo and ask "is this agent-ready?" We deliberately don't. A score you can't reproduce isn't a gate — it's a vibe. Charter's score is a public formula over a fixed rule set: same repo, same score, every time. You can read the rules, predict the result, and trust the number enough to block a merge on it.

## The loop

The workflow is the same three steps whether you're cleaning up one repo or rolling Charter out across an org:

1. `charter init` — scaffold the context files an agent expects.
2. `charter doctor` — get an honest baseline, usually 40–60 on a repo that's never seen this.
3. `charter fix` — apply diff-first repairs for the safe rules, then gate it in CI so the score only ratchets up.

Nothing is written without a diff you approve. Secrets and destructive commands are never auto-touched.

## The contract

Charter makes ten commitments and shows its work on every one — no network, no LLM in the core, no file deletion, no silent mutation, every finding traceable to a rule and a fix, every release signed. It's Apache-2.0 and free, forever.

Run `brew install use-charter/tap/charter` and `charter doctor` on your repo. The first number is usually humbling. Fixing it is the point.
