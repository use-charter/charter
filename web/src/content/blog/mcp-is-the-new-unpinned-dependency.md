---
title: "MCP is the new unpinned dependency"
description: "Agents call MCP servers with the same trust npm calls once got by default — no lockfile, no pin, no allowlist. 2026's supply-chain lesson is repeating one layer up."
date: 2026-07-07
author: "Charter"
tags: ["ai-agents", "security", "mcp"]
---

Ten years ago the lesson was `npm install` and pray. Whatever `left-pad` or `event-stream` did at install time, ran — no review, no pin, no second thought. It took a few incidents and a decade of tooling (lockfiles, `npm audit`, Socket, Snyk) to make that trust look reckless in hindsight.

2026 is repeating the pattern one layer up. Agents now reach for MCP servers the same way — add a server to `.mcp.json`, grant it filesystem or network access, done. No pin, no allowlist, no review of what that server can actually do once the agent starts calling it mid-session.

## The trust model nobody wrote down

An MCP server isn't a library you can `git blame`. It's a live process the agent talks to over stdio or HTTP, and the agent trusts its responses by default — tool descriptions, return values, even embedded instructions in tool output. A malicious or compromised server doesn't need to touch your code; it can just tell the agent to.

That's a bigger blast radius than a bad npm postinstall script, because the thing consuming the payload is a model taking autonomous action, not a build step you can sandbox with a container.

## What "unpinned" looks like in practice

- MCP servers referenced by name or URL, no version pin, no hash
- No allowlist — any server the agent discovers can be called
- Tool descriptions trusted verbatim, never diffed against what shipped last week
- Secrets and filesystem access granted per-session, not scoped per-tool
- No log of which server said what to the agent, so a bad call is invisible until the diff shows up

None of this is exotic. It's the default state of most `.mcp.json` files right now — the same default trust `npm install` had before anyone thought to lock it down.

## The fix is the same shape as last time

Supply-chain security for packages didn't need a breakthrough — it needed boring discipline: lockfiles, pins, scoped permissions, and a scanner that runs before merge, not after an incident. MCP needs the identical discipline, just applied to servers instead of packages:

- pin MCP servers to a known version or hash, not a floating tag
- maintain an explicit allowlist — no server runs unless it's in the repo's config
- scope credentials per-tool, not per-session
- treat tool descriptions and outputs as untrusted input, same as you'd treat a webhook payload

## Where Charter fits

This is exactly what Charter's MCP-safety category checks for — an unpinned server, a missing allowlist, or credentials with more reach than the tool needs, all cap the score the same way a raw secret does. It's a boring, reproducible check, which is the point: the repo shouldn't need an incident to notice its agent trusts every server that shows up.

`charter doctor` flags it before the agent does something with that trust you didn't sign off on. The npm ecosystem needed a few bad years to earn its lockfile habit — MCP doesn't have to.
