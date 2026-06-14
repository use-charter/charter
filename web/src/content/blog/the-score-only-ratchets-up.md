---
title: "The score only ratchets up"
description: "Readiness isn't a one-time cleanup. Here's how Charter turns a repo from an honest baseline into a wall every pull request has to clear — and why the hard caps matter more than the formula."
date: 2026-06-12
author: "Charter"
tags: ["engineering", "philosophy"]
---

Most "health score" tools give you a number once, you feel briefly bad about it, and nothing changes. Charter is built to do the opposite: turn the score into a floor that can't drop.

## The formula is the boring part

The base score is simple and public:

```
score = max(0, 100 − B×20 − H×10 − M×4 − L×1)
```

Blockers cost 20 points, Highs 10, Mediums 4, Lows 1. Informational findings don't deduct. You can do the arithmetic in your head, which is the point — a gate you can't predict is a gate you won't trust.

## The caps are the interesting part

A formula alone can be gamed by a pile of low-severity noise averaging out a real problem. So Charter applies hard caps that override the arithmetic:

- a raw secret in agent-visible content caps the score at **49**
- any active Blocker caps it at **59**

A repo can have a beautiful spread of green categories and still score below passing, because one exposed credential is not something you average away. The cap says: fix this first, nothing else counts until you do.

## Ratcheting in CI

The lifecycle is `init → doctor → fix → gate`. The last step is where the score stops being a vanity metric. Add the GitHub Action with a threshold, and every pull request gets re-scored; anything below the line fails the check, and findings post as SARIF straight into GitHub Code Scanning.

From that point the score can only go up, because regressions can't merge. That's the whole design: not a one-time audit, but a baseline that compounds.

## Stable within a major version

Because teams gate on it, the formula and caps are stable within a major version. We won't silently re-weight a rule and break your build on a Tuesday. When the scoring changes, it's a major version and it's in the changelog — the same contract we ask agents to give your codebase.
