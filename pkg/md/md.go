package md

import (
	"regexp"
	"strings"

	cf "github.com/kentaro-m/blackfriday-confluence"
	bf "github.com/russross/blackfriday/v2"

	"github.com/ankitpokhrel/jira-cli/pkg/md/jirawiki"
)

var jiraWikiCodeSpan = regexp.MustCompile(`\{\{([^{}\n]*)\}\}`)

var jiraInlineCodeUnescaper = strings.NewReplacer(
	`\*`, "*",
	`\_`, "_",
	`\-`, "-",
	`\+`, "+",
	`\^`, "^",
	`\~`, "~",
	`\{`, "{",
	`\!`, "!",
	`\[`, "[",
	`\]`, "]",
	`\(`, "(",
	`\)`, ")",
)

// ToJiraMD translates CommonMark to Jira flavored markdown.
func ToJiraMD(md string) string {
	if md == "" {
		return md
	}

	renderer := &cf.Renderer{Flags: cf.IgnoreMacroEscaping}
	r := bf.New(bf.WithRenderer(renderer), bf.WithExtensions(bf.CommonExtensions))

	return restoreMarkdownCodeSpans(string(renderer.Render(r.Parse([]byte(md)))))
}

// FromJiraMD translates Jira flavored markdown to CommonMark.
func FromJiraMD(jfm string) string {
	return jirawiki.Parse(jfm)
}

func restoreMarkdownCodeSpans(jiraMD string) string {
	return jiraWikiCodeSpan.ReplaceAllStringFunc(jiraMD, func(span string) string {
		return markdownCodeSpan(jiraInlineCodeUnescaper.Replace(span[2 : len(span)-2]))
	})
}

func markdownCodeSpan(text string) string {
	ticks := strings.Repeat("`", longestBacktickRun(text)+1)
	if strings.HasPrefix(text, "`") || strings.HasSuffix(text, "`") {
		return ticks + " " + text + " " + ticks
	}

	return ticks + text + ticks
}

func longestBacktickRun(text string) int {
	longest, current := 0, 0
	for _, r := range text {
		if r == '`' {
			current++
			longest = max(longest, current)
		} else {
			current = 0
		}
	}

	return longest
}
