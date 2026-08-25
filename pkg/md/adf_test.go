package md

import (
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
