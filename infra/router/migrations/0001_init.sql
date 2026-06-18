-- Founder dashboard analytics — initial schema (ADR-0027, spec 2026-06-19).
-- Aggregate, UTC-day-keyed tables. WITHOUT ROWID so the primary key IS the
-- storage (no separate rowid index) — halves rows-written per upsert, which
-- matters against the D1 Free 100k-writes/day budget. No raw IP/UA/query string
-- is ever stored: `vhash` is a daily-salted SHA-256 (salt rotates + is deleted
-- within ~48h via KV), so rows are not reversible to a person.

CREATE TABLE IF NOT EXISTS pageview (
  day  TEXT    NOT NULL,            -- UTC date, YYYY-MM-DD
  path TEXT    NOT NULL,            -- normalized, allow-listed (else '/__other__')
  hits INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (day, path)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS visitor (
  day   TEXT NOT NULL,              -- UTC date
  vhash TEXT NOT NULL,              -- daily-salted SHA-256(IP+UA), base64url, 16 bytes
  PRIMARY KEY (day, vhash)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS geo (
  day     TEXT    NOT NULL,         -- UTC date
  country TEXT    NOT NULL,         -- ISO-2 from request.cf.country (T1/XX excluded)
  hits    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (day, country)
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS event (
  day   TEXT NOT NULL,              -- UTC date
  type  TEXT NOT NULL,             -- allow-listed (e.g. 'install_copied')
  vhash TEXT NOT NULL,             -- daily visitor hash → dedup + spam resistance
  PRIMARY KEY (day, type, vhash)
) WITHOUT ROWID;
