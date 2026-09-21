package activity

import (
	"strings"
	"testing"
)

func TestRenderMarkdownDetailIncludesBody(t *testing.T) {
	section := KindSection{
		Kind:  KindComments,
		Count: 1,
		Entries: []Entry{{
			Title:     "owner/repo#1 Fix the thing",
			URL:       "https://github.com/owner/repo/issues/1#issuecomment-1",
			CreatedAt: "2026-04-01",
			Body:      "first line\n\n# not a heading\nsecond line",
		}},
	}

	got := RenderMarkdown(section, ModeDetail)
	want := strings.Join([]string{
		"# comments",
		"",
		"Count: 1",
		"",
		"## owner/repo#1 Fix the thing",
		"",
		"- Date: 2026-04-01",
		"- URL: <https://github.com/owner/repo/issues/1#issuecomment-1>",
		"",
		"> first line",
		">",
		"> # not a heading",
		"> second line",
		"",
	}, "\n")
	if got != want {
		t.Errorf("RenderMarkdown(detail) =\n%q\nwant\n%q", got, want)
	}
}

func TestRenderMarkdownSummaryOmitsBody(t *testing.T) {
	section := KindSection{
		Kind:  KindComments,
		Count: 1,
		Entries: []Entry{{
			Title:     "owner/repo#1 Fix the thing",
			URL:       "https://github.com/owner/repo/issues/1",
			CreatedAt: "2026-04-01",
			Body:      "body text",
		}},
	}

	got := RenderMarkdown(section, ModeSummary)
	if strings.Contains(got, "body text") {
		t.Errorf("summary mode should not render the body, got:\n%s", got)
	}
	if !strings.Contains(got, "- [owner/repo#1 Fix the thing](https://github.com/owner/repo/issues/1)") {
		t.Errorf("summary mode should render a link line, got:\n%s", got)
	}
}
