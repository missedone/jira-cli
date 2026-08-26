package md

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/jira-cli/pkg/adf"
)

func TestToADF(t *testing.T) {
	input := "Intro with `COMMAND` and `logs/PROJ-123-run.txt`.\n\n" +
		"## Details\n\n" +
		"- **Strong** item with `APP2-42`.\n" +
		"- [Example](https://example.com)"

	expected := &adf.ADF{
		Version: 1,
		DocType: "doc",
		Content: []*adf.Node{
			{
				NodeType: adf.NodeParagraph,
				Content: []*adf.Node{
					{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: "Intro with "}},
					{
						NodeType: adf.ChildNodeText,
						NodeValue: adf.NodeValue{
							Text:  "COMMAND",
							Marks: []adf.MarkNode{{MarkType: adf.MarkCode}},
						},
					},
					{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: " and "}},
					{
						NodeType: adf.ChildNodeText,
						NodeValue: adf.NodeValue{
							Text:  "logs/PROJ-123-run.txt",
							Marks: []adf.MarkNode{{MarkType: adf.MarkCode}},
						},
					},
					{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: "."}},
				},
			},
			{
				NodeType: adf.NodeHeading,
				Attributes: map[string]any{
					"level": 2,
				},
				Content: []*adf.Node{
					{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: "Details"}},
				},
			},
			{
				NodeType: adf.NodeBulletList,
				Content: []*adf.Node{
					{
						NodeType: adf.ChildNodeListItem,
						Content: []*adf.Node{
							{
								NodeType: adf.NodeParagraph,
								Content: []*adf.Node{
									{
										NodeType: adf.ChildNodeText,
										NodeValue: adf.NodeValue{
											Text:  "Strong",
											Marks: []adf.MarkNode{{MarkType: adf.MarkStrong}},
										},
									},
									{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: " item with "}},
									{
										NodeType: adf.ChildNodeText,
										NodeValue: adf.NodeValue{
											Text:  "APP2-42",
											Marks: []adf.MarkNode{{MarkType: adf.MarkCode}},
										},
									},
									{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: "."}},
								},
							},
						},
					},
					{
						NodeType: adf.ChildNodeListItem,
						Content: []*adf.Node{
							{
								NodeType: adf.NodeParagraph,
								Content: []*adf.Node{
									{
										NodeType: adf.ChildNodeText,
										NodeValue: adf.NodeValue{
											Text: "Example",
											Marks: []adf.MarkNode{{
												MarkType: adf.MarkLink,
												Attributes: map[string]any{
													"href": "https://example.com",
												},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, ToADF(input))
}

func TestToADFCodeDropsIncompatibleMarks(t *testing.T) {
	input := "**`bold`** _`em`_ ~~`strike`~~ [`linked`](https://example.com)"

	assertADFJSON(t, `{
		"version": 1,
		"type": "doc",
		"content": [{
			"type": "paragraph",
			"content": [
				{"type": "text", "text": "bold", "marks": [{"type": "code"}]},
				{"type": "text", "text": " "},
				{"type": "text", "text": "em", "marks": [{"type": "code"}]},
				{"type": "text", "text": " "},
				{"type": "text", "text": "strike", "marks": [{"type": "code"}]},
				{"type": "text", "text": " "},
				{"type": "text", "text": "linked", "marks": [{"type": "link", "attrs": {"href": "https://example.com"}}, {"type": "code"}]}
			]
		}]
	}`, ToADF(input))
}

func TestToADFConvertsTable(t *testing.T) {
	input := "| Name | Value |\n" +
		"| --- | --- |\n" +
		"| Command | `COMMAND` |"

	assertADFJSON(t, `{
		"version": 1,
		"type": "doc",
		"content": [{
			"type": "table",
			"attrs": {"isNumberColumnEnabled": false, "layout": "default"},
			"content": [{
				"type": "tableRow",
				"content": [
					{"type": "tableHeader", "content": [{"type": "paragraph", "content": [{"type": "text", "text": "Name"}]}]},
					{"type": "tableHeader", "content": [{"type": "paragraph", "content": [{"type": "text", "text": "Value"}]}]}
				]
			}, {
				"type": "tableRow",
				"content": [
					{"type": "tableCell", "content": [{"type": "paragraph", "content": [{"type": "text", "text": "Command"}]}]},
					{"type": "tableCell", "content": [{"type": "paragraph", "content": [{"type": "text", "text": "COMMAND", "marks": [{"type": "code"}]}]}]}
				]
			}]
		}]
	}`, ToADF(input))
}

func assertADFJSON(t *testing.T, expected string, doc *adf.ADF) {
	t.Helper()

	actual, err := json.Marshal(doc)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(actual))
}
