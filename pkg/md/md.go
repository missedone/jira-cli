package md

import (
	"regexp"
	"strings"

	cf "github.com/kentaro-m/blackfriday-confluence"
	bf "github.com/russross/blackfriday/v2"

	"github.com/ankitpokhrel/jira-cli/pkg/md/jirawiki"
)

var jiraIssueKeyCodeSpan = regexp.MustCompile(`\{\{([^{}\n]*(?:[A-Z][A-Z0-9]+\\?-[0-9]+)[^{}\n]*)\}\}`)

// ToJiraMD translates CommonMark to Jira flavored markdown.
func ToJiraMD(md string) string {
	if md == "" {
		return md
	}

	renderer := &cf.Renderer{Flags: cf.IgnoreMacroEscaping}
	r := bf.New(bf.WithRenderer(renderer), bf.WithExtensions(bf.CommonExtensions))

	return unwrapIssueKeyCodeSpans(string(renderer.Render(r.Parse([]byte(md)))))
}

// FromJiraMD translates Jira flavored markdown to CommonMark.
func FromJiraMD(jfm string) string {
	return jirawiki.Parse(jfm)
}

func unwrapIssueKeyCodeSpans(jiraMD string) string {
	return jiraIssueKeyCodeSpan.ReplaceAllStringFunc(jiraMD, func(span string) string {
		return strings.ReplaceAll(span[2:len(span)-2], `\-`, "-")
	})
}
