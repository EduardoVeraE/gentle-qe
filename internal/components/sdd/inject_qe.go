package sdd

import "strings"

const qeFrontmatterFence = "---"

// qeSwapNativeAgentBody preserves YAML frontmatter (name/model/effort/tools),
// rewrites the description: entry to qeDescription, and replaces the body
// (everything after the closing frontmatter fence) with qeBody.
//
// Defensive no-op if wrapper has no frontmatter fence — Path 2 must never
// corrupt a native-agent wrapper it cannot safely parse.
func qeSwapNativeAgentBody(wrapper, qeBody, qeDescription string) string {
	if !strings.HasPrefix(wrapper, qeFrontmatterFence+"\n") {
		return wrapper
	}

	afterOpen := wrapper[len(qeFrontmatterFence)+1:]
	closeMarker := "\n" + qeFrontmatterFence
	closeRel := strings.Index(afterOpen, closeMarker)
	if closeRel == -1 {
		return wrapper // no closing fence found — defensive no-op
	}

	frontmatterBody := afterOpen[:closeRel]
	newFrontmatterBody := qeRewriteDescription(frontmatterBody, qeDescription)

	var b strings.Builder
	b.WriteString(qeFrontmatterFence)
	b.WriteString("\n")
	b.WriteString(newFrontmatterBody)
	b.WriteString("\n")
	b.WriteString(qeFrontmatterFence)
	b.WriteString("\n\n")
	b.WriteString(strings.TrimLeft(qeStripOwnFrontmatter(qeBody), "\n"))
	return b.String()
}

// qeStripOwnFrontmatter removes a QE asset's own YAML frontmatter block (the
// fenced `---\n...\n---\n` header every _qe-sdd/{phase}.md file carries) so
// only the real body content is inserted into the native sub-agent wrapper.
// Without this, qeSwapNativeAgentBody would paste the QE asset's frontmatter
// in raw as body text, producing a wrapper with two frontmatter blocks.
//
// Uses the same fence-detection logic as qeSwapNativeAgentBody itself. If
// qeBody has no leading frontmatter fence, it is returned unchanged — fail-
// open for QE assets that don't carry their own frontmatter.
func qeStripOwnFrontmatter(qeBody string) string {
	if !strings.HasPrefix(qeBody, qeFrontmatterFence+"\n") {
		return qeBody
	}

	afterOpen := qeBody[len(qeFrontmatterFence)+1:]
	closeMarker := "\n" + qeFrontmatterFence
	closeRel := strings.Index(afterOpen, closeMarker)
	if closeRel == -1 {
		return qeBody // no closing fence found — defensive no-op
	}

	// closeRel points to the start of "\n---"; the real body starts right
	// after that fence line's own trailing newline.
	afterClose := afterOpen[closeRel+len(closeMarker):]
	afterClose = strings.TrimPrefix(afterClose, "\n")
	return afterClose
}

// qeRewriteDescription replaces the description: entry inside a frontmatter
// block (the lines between the fences) with a single-line QE description,
// preserving every other line (name/model/effort/tools) byte-identical and
// in their original order. Handles both single-line (`description: "..."`)
// and folded/literal block scalar forms (`description: >` / `description: |`
// followed by indented continuation lines).
func qeRewriteDescription(frontmatterBody, qeDescription string) string {
	lines := strings.Split(frontmatterBody, "\n")
	out := make([]string, 0, len(lines))
	replaced := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !replaced && strings.HasPrefix(strings.TrimLeft(line, " "), "description:") {
			out = append(out, "description: "+qeDescription)
			replaced = true
			i++
			// Consume folded/literal block scalar continuation lines: any
			// subsequent line indented with a leading space/tab belongs to
			// the same description value, not a new frontmatter key. A blank
			// line does NOT end the continuation by itself — YAML folded
			// scalars allow blank lines between indented paragraphs — so we
			// only stop once a run of blank lines is followed by a
			// non-indented line (a genuine new frontmatter key or EOF).
			for i < len(lines) {
				if isIndentedContinuation(lines[i]) {
					i++
					continue
				}
				if lines[i] == "" {
					lookahead := i
					for lookahead < len(lines) && lines[lookahead] == "" {
						lookahead++
					}
					if lookahead < len(lines) && isIndentedContinuation(lines[lookahead]) {
						i = lookahead
						continue
					}
				}
				break
			}
			i-- // compensate for the outer loop's own i++ on next iteration
			continue
		}
		out = append(out, line)
	}

	if !replaced {
		// Defensive: no description key found — append one rather than
		// silently dropping the QE description.
		out = append(out, "description: "+qeDescription)
	}

	return strings.Join(out, "\n")
}

// isIndentedContinuation reports whether line is a continuation line of a
// folded/literal YAML block scalar (indented with a leading space or tab).
// A blank line never starts a new top-level frontmatter key on its own, but
// treating only genuinely indented lines as continuations is sufficient for
// every wrapper shape this fork generates.
func isIndentedContinuation(line string) bool {
	if line == "" {
		return false
	}
	return line[0] == ' ' || line[0] == '\t'
}

// qeSubAgentDescription returns the QE-framed one-line frontmatter
// description for a native sub-agent wrapper, used by Path 2's body-swap.
func qeSubAgentDescription(phase string) string {
	switch phase {
	case "sdd-explore":
		return `"Explore SDD ideas as a quality-risk analysis (risk, testability, defect clustering, oracle inventory). Trigger: SDD explore phase for a QE test-design change."`
	case "sdd-propose":
		return `"Create an SDD proposal that reframes capabilities as test-requirements, candidate oracles, and risk statements. Trigger: SDD propose phase for a QE test-design change."`
	case "sdd-spec":
		return `"Write oracle-first SDD delta specs with GIVEN/WHEN/THEN scenarios. Trigger: SDD spec phase for a QE test-design change."`
	case "sdd-design":
		return `"Produce the SDD Testing Strategy: test levels, named ISTQB techniques, test pyramid, and risk-based prioritization. Trigger: SDD design phase for a QE test-design change."`
	case "sdd-tasks":
		return `"Enumerate automatable GIVEN/WHEN/THEN scenarios grouped by test level. Trigger: SDD tasks phase for a QE test-design change."`
	case "sdd-apply":
		return `"Write test code (specs, fixtures, page objects) for the scenarios assigned by sdd-tasks. Trigger: SDD apply phase for a QE test-design change."`
	case "sdd-verify":
		return `"Execute the test suite and prove coverage-by-risk and flakiness. Trigger: SDD verify phase for a QE test-design change."`
	default:
		return `"QE test-design phase for a Gentle-QE SDD change."`
	}
}
