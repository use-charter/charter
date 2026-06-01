// Package sarif renders a doctor.Result as a SARIF 2.1.0 log for GitHub Code
// Scanning. It is a pure projection of the (already redacted) result: no disk
// I/O, deterministic ordering, position-based fingerprints.
package sarif

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/rules/catalog"
	"go.use-charter.dev/charter/internal/version"
)

const schemaURI = "https://json.schemastore.org/sarif-2.1.0.json"

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []runEntry `json:"runs"`
}

type runEntry struct {
	Tool              toolEntry              `json:"tool"`
	AutomationDetails automationDetailsEntry `json:"automationDetails"`
	Results           []resultEntry          `json:"results"`
}

type automationDetailsEntry struct {
	ID string `json:"id"`
}

type toolEntry struct {
	Driver driverEntry `json:"driver"`
}

type driverEntry struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Version        string      `json:"version"`
	Rules          []ruleEntry `json:"rules"`
}

type ruleEntry struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name,omitempty"`
	ShortDescription     textBlock      `json:"shortDescription"`
	HelpURI              string         `json:"helpUri,omitempty"`
	DefaultConfiguration defaultConfig  `json:"defaultConfiguration"`
	Properties           ruleProperties `json:"properties"`
}

type ruleProperties struct {
	Category         string   `json:"category,omitempty"`
	Severity         string   `json:"severity,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	SecuritySeverity string   `json:"security-severity,omitempty"`
}

type defaultConfig struct {
	Level string `json:"level"`
}

type textBlock struct {
	Text string `json:"text"`
}

type resultEntry struct {
	RuleID              string             `json:"ruleId"`
	RuleIndex           int                `json:"ruleIndex"`
	Level               string             `json:"level"`
	Kind                string             `json:"kind,omitempty"`
	Message             textBlock          `json:"message"`
	Locations           []locationEntry    `json:"locations,omitempty"`
	PartialFingerprints map[string]string  `json:"partialFingerprints"`
	Suppressions        []suppressionEntry `json:"suppressions"`
}

type locationEntry struct {
	PhysicalLocation physicalLocationEntry `json:"physicalLocation"`
}

type physicalLocationEntry struct {
	ArtifactLocation artifactLocationEntry `json:"artifactLocation"`
	Region           *regionEntry          `json:"region,omitempty"`
}

type artifactLocationEntry struct {
	URI string `json:"uri"`
}

type regionEntry struct {
	StartLine int `json:"startLine"`
}

type suppressionEntry struct {
	Kind string `json:"kind"`
}

type item struct {
	f          findings.Finding
	suppressed bool
	source     string
}

// Render projects a doctor.Result into a SARIF 2.1.0 log.
func Render(res doctor.Result) ([]byte, error) {
	items := collect(res)
	rules, idx := buildRules(items)
	log := sarifLog{
		Schema:  schemaURI,
		Version: "2.1.0",
		Runs: []runEntry{{
			Tool: toolEntry{Driver: driverEntry{
				Name:           "Charter",
				InformationURI: "https://use-charter.dev",
				Version:        version.Version(),
				Rules:          rules,
			}},
			AutomationDetails: automationDetailsEntry{ID: "charter"},
			Results:           buildResults(items, idx),
		}},
	}
	return json.Marshal(log)
}

func collect(res doctor.Result) []item {
	items := make([]item, 0, len(res.Findings)+len(res.Suppressed))
	for _, f := range res.Findings {
		items = append(items, item{f: f})
	}
	for _, s := range res.Suppressed {
		items = append(items, item{f: s.Finding, suppressed: true, source: s.Source})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if wi, wj := items[i].f.Severity.Weight(), items[j].f.Severity.Weight(); wi != wj {
			return wi > wj
		}
		return items[i].f.RuleID < items[j].f.RuleID
	})
	return items
}

func buildRules(items []item) ([]ruleEntry, map[string]int) {
	levelByRule := map[string]string{}
	severityByRule := map[string]string{}
	infoByRule := map[string]bool{}
	var order []string
	for _, it := range items {
		if _, ok := levelByRule[it.f.RuleID]; ok {
			continue
		}
		levelByRule[it.f.RuleID] = levelFor(it.f)
		severityByRule[it.f.RuleID] = string(it.f.Severity)
		infoByRule[it.f.RuleID] = it.f.Informational
		order = append(order, it.f.RuleID)
	}
	sort.Strings(order)

	idx := make(map[string]int, len(order))
	rules := make([]ruleEntry, 0, len(order))
	for i, id := range order {
		idx[id] = i
		name, desc, helpURI, category := id, id, "", ""
		if e, ok := catalog.Lookup(id); ok {
			name, desc, helpURI, category = e.Name, e.ShortDescription, e.HelpURI, e.Category
		}
		props := ruleProperties{Category: category, Severity: severityByRule[id]}
		if !infoByRule[id] {
			props.Tags = []string{"security"}
			props.SecuritySeverity = securitySeverity(severityByRule[id])
		}
		rules = append(rules, ruleEntry{
			ID:                   id,
			Name:                 name,
			ShortDescription:     textBlock{Text: desc},
			HelpURI:              helpURI,
			DefaultConfiguration: defaultConfig{Level: levelByRule[id]},
			Properties:           props,
		})
	}
	return rules, idx
}

func buildResults(items []item, idx map[string]int) []resultEntry {
	results := make([]resultEntry, 0, len(items))
	for _, it := range items {
		r := resultEntry{
			RuleID:              it.f.RuleID,
			RuleIndex:           idx[it.f.RuleID],
			Level:               levelFor(it.f),
			Message:             textBlock{Text: it.f.Summary},
			Locations:           locationsFor(it.f),
			PartialFingerprints: map[string]string{"primaryLocationLineHash": fingerprint(it.f)},
			Suppressions:        []suppressionEntry{},
		}
		if it.f.Informational {
			r.Kind = "informational"
		}
		if it.suppressed {
			r.Suppressions = []suppressionEntry{{Kind: it.source}}
		}
		results = append(results, r)
	}
	return results
}

func locationsFor(f findings.Finding) []locationEntry {
	var locs []locationEntry
	for _, l := range f.Locations {
		if l.Path == "" {
			continue
		}
		pl := physicalLocationEntry{ArtifactLocation: artifactLocationEntry{URI: l.Path}}
		if l.Line > 0 {
			pl.Region = &regionEntry{StartLine: l.Line}
		}
		locs = append(locs, locationEntry{PhysicalLocation: pl})
	}
	return locs
}

func levelFor(f findings.Finding) string {
	if f.Informational {
		return "note"
	}
	switch f.Severity {
	case findings.SeverityBlocker, findings.SeverityHigh:
		return "error"
	case findings.SeverityMedium:
		return "warning"
	default:
		return "note"
	}
}

// securitySeverity maps a Charter severity to a GitHub Code Scanning
// security-severity number (CVSS-like: critical >=9, high 7-8.9, medium 4-6.9,
// low 0.1-3.9).
func securitySeverity(severity string) string {
	switch findings.Severity(severity) {
	case findings.SeverityBlocker:
		return "9.5"
	case findings.SeverityHigh:
		return "7.5"
	case findings.SeverityMedium:
		return "5.0"
	default:
		return "2.0"
	}
}

// fingerprint is a deterministic, position-based primaryLocationLineHash computed
// purely from the finding (no source I/O).
func fingerprint(f findings.Finding) string {
	h := sha256.New()
	if len(f.Locations) > 0 {
		l := f.Locations[0]
		_, _ = fmt.Fprintf(h, "%s\x00%s\x00%d", f.RuleID, l.Path, l.Line)
	} else {
		_, _ = io.WriteString(h, f.RuleID)
	}
	return hex.EncodeToString(h.Sum(nil))
}
