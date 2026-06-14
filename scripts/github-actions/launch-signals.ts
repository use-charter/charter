#!/usr/bin/env bun
/**
 * Launch-signal monitor (launch-checklist §7).
 *
 * Runs the launch-signal searches against the GitHub search API and reports new
 * hits into a single tracking issue (label `launch-signals`): adoption of the
 * GitHub Action, references to the published charter.yaml schema, and unprompted
 * "charter doctor" / "use-charter.dev" mentions across code and issues.
 *
 * State (the keys already reported) lives in a marker inside the issue body, so
 * each weekly run only surfaces what is genuinely new — no committed state, no
 * external service. GitHub search only indexes public repositories, so this is
 * meaningful once the repo is public; before then every query simply returns
 * nothing.
 *
 * Auth: GITHUB_TOKEN (issues: write). Repo: GITHUB_REPOSITORY (owner/repo).
 */

const API = "https://api.github.com";
const TOKEN = process.env["GITHUB_TOKEN"] ?? "";
const REPO = process.env["GITHUB_REPOSITORY"] ?? "";
const LABEL = "launch-signals";
const STATE_RE = /<!--launch-signals-state:(.*?)-->/s;
const MAX_SEEN_PER_SIGNAL = 400; // bound issue-body growth
const PER_QUERY = 30; // top-N results per query

const [OWNER, NAME] = REPO.split("/");
if (!TOKEN || !OWNER || !NAME) {
  console.error("GITHUB_TOKEN and GITHUB_REPOSITORY (owner/repo) are required");
  process.exit(2);
}

type SearchKind = "code" | "issues";
interface Signal {
  key: string;
  title: string;
  kind: SearchKind;
  query: string;
}

// The repo's own org is excluded so we don't count ourselves as an adopter.
const EXCLUDE_OWNER = OWNER;

const SIGNALS: Signal[] = [
  { key: "adoption", title: "Signal 1 — CI adoption (`use-charter/charter-action`)", kind: "code", query: `"use-charter/charter-action"` },
  { key: "schema", title: "Config adoption (charter.yaml `$schema`)", kind: "code", query: `"use-charter.dev/schema/charter.schema.json"` },
  { key: "doctor-code", title: 'Signal 3 — "charter doctor" in code', kind: "code", query: `"charter doctor"` },
  { key: "doctor-issues", title: 'Signal 3 — "charter doctor" in issues / PRs', kind: "issues", query: `"charter doctor"` },
  { key: "domain", title: "`use-charter.dev` references in code", kind: "code", query: `"use-charter.dev"` },
];

interface Hit {
  key: string; // stable dedupe key
  label: string; // human display
  url: string;
}

interface State {
  seen: Record<string, string[]>;
  lastRun: string | null;
}

async function gh<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path.startsWith("http") ? path : `${API}${path}`, {
    ...init,
    headers: {
      Authorization: `Bearer ${TOKEN}`,
      Accept: "application/vnd.github+json",
      "X-GitHub-Api-Version": "2022-11-28",
      "User-Agent": "charter-launch-signals",
      ...(init?.headers ?? {}),
    },
  });
  if (!res.ok) {
    throw new Error(`${init?.method ?? "GET"} ${path} → ${res.status} ${await res.text()}`);
  }
  return (await res.json()) as T;
}

interface CodeItem {
  path: string;
  html_url: string;
  repository: { full_name: string; owner: { login: string } };
}
interface IssueItem {
  title: string;
  html_url: string;
  repository_url: string;
  user: { login: string };
}

async function search(signal: Signal): Promise<Hit[]> {
  const endpoint = signal.kind === "code" ? "/search/code" : "/search/issues";
  const url = `${endpoint}?q=${encodeURIComponent(signal.query)}&per_page=${PER_QUERY}&sort=indexed&order=desc`;
  const data = await gh<{ items: (CodeItem | IssueItem)[] }>(url);
  const hits: Hit[] = [];
  for (const it of data.items ?? []) {
    if (signal.kind === "code") {
      const c = it as CodeItem;
      if (c.repository.owner.login.toLowerCase() === EXCLUDE_OWNER.toLowerCase()) continue;
      hits.push({ key: `${c.repository.full_name}:${c.path}`, label: `${c.repository.full_name} · \`${c.path}\``, url: c.html_url });
    } else {
      const i = it as IssueItem;
      const repo = i.repository_url.replace(`${API}/repos/`, "");
      if (repo.split("/")[0]?.toLowerCase() === EXCLUDE_OWNER.toLowerCase()) continue;
      hits.push({ key: i.html_url, label: `${repo} · ${i.title}`, url: i.html_url });
    }
  }
  return hits;
}

async function findOrCreateIssue(): Promise<{ number: number; body: string }> {
  const issues = await gh<{ number: number; body: string | null }[]>(
    `/repos/${OWNER}/${NAME}/issues?labels=${LABEL}&state=open&per_page=1`,
  );
  const existing = issues[0];
  if (existing) return { number: existing.number, body: existing.body ?? "" };

  // Ensure the label exists (ignore "already exists").
  try {
    await gh(`/repos/${OWNER}/${NAME}/labels`, {
      method: "POST",
      body: JSON.stringify({ name: LABEL, color: "5aa2f7", description: "Automated launch-signal monitor (§7)" }),
    });
  } catch {
    /* label already exists */
  }
  const created = await gh<{ number: number; body: string }>(`/repos/${OWNER}/${NAME}/issues`, {
    method: "POST",
    body: JSON.stringify({
      title: "📡 Launch signals",
      labels: [LABEL],
      body: `Automated weekly digest of Charter launch signals — adoption, mentions, and references found via GitHub search. New hits are posted as comments below.\n\n<!--launch-signals-state:{"seen":{},"lastRun":null}-->`,
    }),
  });
  return { number: created.number, body: created.body ?? "" };
}

function parseState(body: string): State {
  const m = body.match(STATE_RE);
  if (m?.[1]) {
    try {
      return JSON.parse(m[1]) as State;
    } catch {
      /* fall through to fresh state */
    }
  }
  return { seen: {}, lastRun: null };
}

function renderBody(state: State): string {
  return [
    "Automated weekly digest of Charter launch signals — adoption, mentions, and references found via GitHub search. New hits are posted as comments below.",
    "",
    `_Last run: ${state.lastRun ?? "never"}._`,
    "",
    `<!--launch-signals-state:${JSON.stringify(state)}-->`,
  ].join("\n");
}

async function main(): Promise<void> {
  const issue = await findOrCreateIssue();
  const state = parseState(issue.body);

  const sections: string[] = [];
  const errors: string[] = [];
  let totalNew = 0;

  for (const signal of SIGNALS) {
    let hits: Hit[] = [];
    try {
      hits = await search(signal);
    } catch (err) {
      errors.push(`- ${signal.title}: search failed — ${err instanceof Error ? err.message : String(err)}`);
      continue;
    }
    const seen = new Set(state.seen[signal.key] ?? []);
    const fresh = hits.filter((h) => !seen.has(h.key));
    if (fresh.length > 0) {
      totalNew += fresh.length;
      sections.push(`### ${signal.title} — ${fresh.length} new`);
      for (const h of fresh.slice(0, 25)) sections.push(`- [${h.label}](${h.url})`);
      sections.push("");
    }
    // Merge fresh into seen, bounded (most-recent-first).
    const merged = [...hits.map((h) => h.key), ...(state.seen[signal.key] ?? [])];
    state.seen[signal.key] = [...new Set(merged)].slice(0, MAX_SEEN_PER_SIGNAL);
  }

  // All queries errored → fail loudly (likely auth/rate problem).
  if (errors.length === SIGNALS.length) {
    console.error(`All ${SIGNALS.length} signal queries failed:\n${errors.join("\n")}`);
    process.exit(1);
  }

  state.lastRun = new Date().toISOString();
  await gh(`/repos/${OWNER}/${NAME}/issues/${issue.number}`, {
    method: "PATCH",
    body: JSON.stringify({ body: renderBody(state) }),
  });

  if (totalNew > 0 || errors.length > 0) {
    const comment = [
      totalNew > 0 ? `**${totalNew} new launch signal${totalNew === 1 ? "" : "s"}** this run:` : "_No new signals this run._",
      "",
      ...sections,
      errors.length > 0 ? `<details><summary>${errors.length} query error(s)</summary>\n\n${errors.join("\n")}\n</details>` : "",
    ].join("\n");
    await gh(`/repos/${OWNER}/${NAME}/issues/${issue.number}/comments`, {
      method: "POST",
      body: JSON.stringify({ body: comment }),
    });
  }

  console.log(`launch-signals: ${totalNew} new, ${errors.length} errors, issue #${issue.number}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
