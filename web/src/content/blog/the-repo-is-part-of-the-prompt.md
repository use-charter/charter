---
title: "The repo is part of the prompt"
description: "Coding agents inherit your repository — its docs, conventions, setup, and gaps. When they fail, the repo is often the reason. Why agent-readiness belongs in CI, not in vibes."
date: 2026-06-24
author: "Charter"
tags: ["ai-agents", "philosophy"]
---

Most conversations about coding agents turn into model comparisons. Which one writes cleaner code? Which one follows instructions? Which one handles a bigger repo? Fair questions — but they skip past the part that decides most outcomes.

An agent never lands in a clean, abstract problem. It lands in **your repository**. And the repository is part of the prompt.

## What the agent inherits

The moment an agent starts working, it absorbs everything around the task: the folder structure, the naming conventions, the missing docs, the stale setup steps, the test suite that may or may not run, the config nobody remembers writing, the internal patterns only three people understand. Then we hand it a vague task and ask for a safe change.

When it gets something wrong, we blame the model. Sometimes that's right. Often the repo just gave it a bad place to work.

## "The model was bad" is usually the wrong diagnosis

The failures look like intelligence problems but rarely are:

- it rewrites far more than it should
- it ignores a project convention
- it adds a dependency the team would never approve
- it can't figure out how to run the tests
- it edits files it should have left alone
- it follows setup instructions that stopped working months ago

Drop a *human* engineer into that same repo — no onboarding doc, no architecture notes, no contribution guide, no clear test command — and we wouldn't call them a bad engineer on day one. We'd say the repo has poor onboarding. Agents have the same problem. They just hit it faster, and with more confidence.

## Good agent workflows are boring

The repos where agents behave aren't magical. They're boringly well-kept:

- a clear `AGENTS.md` (or equivalent) instruction file
- documented setup, pinned tools and versions
- an obvious, discoverable test command
- CI that returns useful feedback
- secrets and dangerous commands behind clear boundaries
- enough written context that "good" is inferable, not guessed

None of that makes the agent perfect. It changes the *failure mode* — from wild guessing to a bounded change that runs the expected checks and explains itself. That's the difference between rails and a cliff.

## Readiness should be measurable

We already measure the rest of software quality — coverage, lint, build status, dependency scans, secret scans, CI health. For agent-readiness we still use vibes: *"Claude Code struggles here," "Cursor is great in this repo but weird in that one."* Useful, but not operational.

Better questions are answerable:

- Does the repo explain its own conventions?
- Can an agent find the correct setup and test path?
- Are tool versions pinned?
- Are dangerous commands documented or isolated?
- Are secrets protected from a casual read?
- Is there enough context to avoid broad guessing?

And the one that matters most: **can this be checked again as the repo changes?** Readiness isn't a one-time cleanup. It should improve, regress, and gate in CI like anything else.

## Where Charter fits

This is the gap Charter is built for. It's an offline CLI that gives a repository a deterministic `0–100` readiness score across the areas that decide whether an agent succeeds — context, secrets, MCP/tool safety, environment, CI, tests, and governance — then points at the concrete gaps. No LLM scoring, no network calls: same repo, same score, every time. Gate it in CI and the number stops being a vanity metric and becomes a floor.

## The model matters. The repo matters too.

Better models are coming. They'll reason better and miss less. But they'll still inherit the environment we hand them. If the repo is confusing, unsafe, undocumented, and hard to verify, even a strong agent can make things worse.

So when an agent keeps failing in a repo, it's worth asking the unflattering question: maybe the repo is part of the problem.

`brew install use-charter/tap/charter && charter doctor` — the first number is usually humbling. Fixing it is the point.
