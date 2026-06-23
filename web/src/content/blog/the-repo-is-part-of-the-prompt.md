---
title: "The repo is part of the prompt"
description: "A coding agent never works on an abstract problem — it works in your repository, and inherits every gap in it. When agents fail, the repo is often the reason. Why agent-readiness belongs in CI."
date: 2026-06-24
author: "Charter"
tags: ["ai-agents", "philosophy"]
---

Every coding-agent conversation collapses into the same debate: which model is smartest, cleanest, most obedient. It's the wrong axis.

An agent never works on an abstract problem. It works inside *your repository* — and the repository is part of the prompt.

## What the agent inherits

The moment it starts, the agent inherits everything around the task: your folder layout and naming, the missing docs, the stale setup steps, a test suite that may not run, config only a few people understand.

Then we hand it a vague task, ask for a safe change, and blame the model when it slips. Sometimes that's fair. More often, the repo gave it nowhere good to stand.

## "The model was bad" is usually the wrong diagnosis

The failures look like intelligence problems. They rarely are:

- it rewrites far more than it should
- it ignores a project convention
- it adds a dependency the team would never approve
- it can't figure out how to run the tests
- it edits files it should have left alone
- it follows setup steps that broke months ago

Drop a *human* engineer into that same repo — no onboarding doc, no architecture notes, no contribution guide, no clear way to run tests — and we wouldn't call them a bad engineer on day one. We'd say the onboarding is broken.

Agents have the same problem. They just hit it faster, and with more confidence.

## Good agent workflows are boring

The repos where agents behave aren't clever. They're boringly well-kept:

- a clear `AGENTS.md` the agent can actually find
- documented setup, with tools and versions pinned
- one obvious, discoverable test command
- CI that returns feedback worth reading
- secrets and destructive commands behind clear walls
- enough written context that "good" is inferable, not guessed

None of that makes the agent perfect. It changes the *failure mode* — from wild guessing to a bounded change that runs the expected checks and explains itself. Rails instead of a cliff.

## Readiness should be measurable

We already measure the rest of software quality: coverage, lint, build status, dependency and secret scans. For agent-readiness we still trade in vibes — *Claude Code struggles here; Cursor's great in this repo, weird in that one.*

Useful, but not operational. Better questions have answers:

- Does the repo state its own conventions?
- Can an agent find the setup and test path?
- Are tool versions pinned?
- Are dangerous commands documented or isolated?
- Are secrets safe from a casual read?
- Is there enough context to stop the guessing?

And the one that matters most: can you check it again as the repo changes? Readiness isn't a one-time cleanup. It should improve, regress, and gate in CI — like everything else that decides whether code ships.

## Where Charter fits

That's the gap Charter is built for. It's an offline CLI that scores a repository `0–100` on agent-readiness — across context, secrets, MCP/tool safety, environment, CI, tests, and governance — then points at the exact gaps.

No LLM scoring, no network calls: same repo, same score, every time. Gate it in CI and the number stops being a vanity metric and becomes a floor a regression can't sink below.

## The model matters. The repo matters too.

Better models are coming. They'll reason harder and miss less. They'll still inherit the environment we hand them — and a confusing, unsafe, undocumented repo can make even a strong agent dangerous.

So when an agent keeps failing in one repo, it's worth asking the unflattering question: maybe the repo is the problem.

`brew install use-charter/tap/charter && charter doctor` — the first number usually stings. Closing the gap is the point.
