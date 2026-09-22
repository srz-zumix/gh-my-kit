package activity

import (
	"fmt"
	"strings"
)

// RenderMarkdown renders a KindSection as Markdown. In ModeSummary, each
// entry is rendered as a single link line; in ModeDetail, every entry becomes
// its own section with its link, creation time and body text.
func RenderMarkdown(section KindSection, mode Mode) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", section.Kind)
	fmt.Fprintf(&b, "Count: %d\n\n", section.Count)

	if mode == ModeDetail {
		for _, e := range section.Entries {
			writeDetailEntry(&b, e)
		}
	} else {
		for _, e := range section.Entries {
			fmt.Fprintf(&b, "- %s\n", entryLine(e))
		}
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// writeDetailEntry renders a single entry as a section with its metadata and
// body.
func writeDetailEntry(b *strings.Builder, e Entry) {
	title := strings.TrimSpace(strings.ReplaceAll(entryTitle(e), "\n", " "))
	fmt.Fprintf(b, "## %s\n\n", title)

	meta := false
	if e.CreatedAt != "" {
		fmt.Fprintf(b, "- Date: %s\n", e.CreatedAt)
		meta = true
	}
	if e.URL != "" {
		fmt.Fprintf(b, "- URL: <%s>\n", e.URL)
		meta = true
	}
	if meta {
		b.WriteString("\n")
	}

	if e.Body != "" {
		b.WriteString(quoteBody(e.Body))
		b.WriteString("\n")
	}
}

// quoteBody renders body as a Markdown blockquote so that its own headings and
// lists cannot break the surrounding document structure.
func quoteBody(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		if line == "" {
			b.WriteString(">\n")
			continue
		}
		fmt.Fprintf(&b, "> %s\n", line)
	}
	return b.String()
}

func entryTitle(e Entry) string {
	if e.Title == "" {
		return "(no title)"
	}
	return e.Title
}

func entryLine(e Entry) string {
	title := entryTitle(e)
	if e.URL == "" {
		return title
	}
	return fmt.Sprintf("[%s](%s)", title, e.URL)
}

// RenderSummaryOverview renders a top-level Markdown overview listing the
// count for every collected kind, at the root and for each owner.
func RenderSummaryOverview(r *Result, kinds []Kind) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Activity summary for %s\n\n", r.Username)
	fmt.Fprintf(&b, "Period: %s .. %s\n\n", r.Since.Format(dateLayout), r.Until.Format(dateLayout))

	fmt.Fprintf(&b, "## Overview\n\n")
	for _, kind := range kinds {
		if !isRootKind(kind) {
			continue
		}
		section := r.Section(kind)
		fmt.Fprintf(&b, "- %s: %d\n", kind, section.Count)
	}
	b.WriteString("\n")

	for _, owner := range r.OwnerLogins() {
		oa := r.Owners[owner]
		fmt.Fprintf(&b, "## %s\n\n", owner)
		for _, kind := range kinds {
			if isRootKind(kind) {
				continue
			}
			section := oa.Section(kind)
			fmt.Fprintf(&b, "- %s: %d\n", kind, section.Count)
		}
		b.WriteString("\n")
	}

	if len(r.Warnings) > 0 {
		b.WriteString("## Warnings\n\n")
		for _, w := range r.Warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
	}

	return b.String()
}
