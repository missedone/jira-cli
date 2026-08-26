package md

import (
	"strings"

	bf "github.com/russross/blackfriday/v2"

	"github.com/ankitpokhrel/jira-cli/pkg/adf"
)

// ToADF translates CommonMark to Atlassian Document Format.
func ToADF(markdown string) *adf.ADF {
	parser := bf.New(bf.WithExtensions(bf.CommonExtensions))
	root := parser.Parse([]byte(markdown))

	doc := &adf.ADF{
		Version: 1,
		DocType: "doc",
		Content: make([]*adf.Node, 0),
	}

	for child := root.FirstChild; child != nil; child = child.Next {
		doc.Content = append(doc.Content, toADFBlocks(child)...)
	}

	if len(doc.Content) == 0 {
		doc.Content = append(doc.Content, paragraph(nil))
	}

	return doc
}

func toADFBlocks(node *bf.Node) []*adf.Node {
	switch node.Type {
	case bf.Heading:
		return []*adf.Node{{
			NodeType: adf.NodeHeading,
			Attributes: map[string]any{
				"level": node.Level,
			},
			Content: toADFInlines(node, nil),
		}}
	case bf.Paragraph:
		return []*adf.Node{paragraph(toADFInlines(node, nil))}
	case bf.List:
		return []*adf.Node{toADFList(node)}
	case bf.Table:
		return []*adf.Node{toADFTable(node)}
	case bf.Item:
		return []*adf.Node{toADFListItem(node)}
	case bf.BlockQuote:
		return []*adf.Node{{
			NodeType: adf.NodeBlockquote,
			Content:  toADFChildBlocks(node),
		}}
	case bf.CodeBlock:
		return []*adf.Node{{
			NodeType:   adf.NodeCodeBlock,
			Attributes: codeBlockAttrs(node),
			Content:    []*adf.Node{textNode(strings.TrimRight(string(node.Literal), "\n"), nil)},
		}}
	default:
		if node.FirstChild != nil {
			return toADFChildBlocks(node)
		}
	}

	return nil
}

func toADFChildBlocks(node *bf.Node) []*adf.Node {
	blocks := make([]*adf.Node, 0)
	for child := node.FirstChild; child != nil; child = child.Next {
		blocks = append(blocks, toADFBlocks(child)...)
	}
	return blocks
}

func toADFList(node *bf.Node) *adf.Node {
	nodeType := adf.NodeBulletList
	if node.ListFlags&bf.ListTypeOrdered != 0 {
		nodeType = adf.NodeOrderedList
	}

	items := make([]*adf.Node, 0)
	for child := node.FirstChild; child != nil; child = child.Next {
		if child.Type == bf.Item {
			items = append(items, toADFListItem(child))
		}
	}

	return &adf.Node{NodeType: nodeType, Content: items}
}

func toADFListItem(node *bf.Node) *adf.Node {
	return &adf.Node{
		NodeType: adf.ChildNodeListItem,
		Content:  toADFChildBlocks(node),
	}
}

func toADFTable(node *bf.Node) *adf.Node {
	rows := make([]*adf.Node, 0)
	for child := node.FirstChild; child != nil; child = child.Next {
		switch child.Type {
		case bf.TableHead, bf.TableBody:
			for row := child.FirstChild; row != nil; row = row.Next {
				if row.Type == bf.TableRow {
					rows = append(rows, toADFTableRow(row))
				}
			}
		case bf.TableRow:
			rows = append(rows, toADFTableRow(child))
		}
	}

	return &adf.Node{
		NodeType: adf.NodeTable,
		Attributes: map[string]any{
			"isNumberColumnEnabled": false,
			"layout":                "default",
		},
		Content: rows,
	}
}

func toADFTableRow(node *bf.Node) *adf.Node {
	cells := make([]*adf.Node, 0)
	for child := node.FirstChild; child != nil; child = child.Next {
		if child.Type == bf.TableCell {
			cells = append(cells, toADFTableCell(child))
		}
	}

	return &adf.Node{NodeType: adf.ChildNodeTableRow, Content: cells}
}

func toADFTableCell(node *bf.Node) *adf.Node {
	nodeType := adf.ChildNodeTableCell
	if node.IsHeader {
		nodeType = adf.ChildNodeTableHeader
	}

	content := toADFChildBlocks(node)
	if len(content) == 0 {
		content = []*adf.Node{paragraph(toADFInlines(node, nil))}
	}

	return &adf.Node{NodeType: nodeType, Content: content}
}

func paragraph(content []*adf.Node) *adf.Node {
	return &adf.Node{
		NodeType: adf.NodeParagraph,
		Content:  content,
	}
}

func codeBlockAttrs(node *bf.Node) map[string]any {
	language := strings.TrimSpace(string(node.Info))
	if language == "" {
		return nil
	}

	return map[string]any{"language": strings.Fields(language)[0]}
}

func toADFInlines(node *bf.Node, marks []adf.MarkNode) []*adf.Node {
	inlines := make([]*adf.Node, 0)

	for child := node.FirstChild; child != nil; child = child.Next {
		switch child.Type {
		case bf.Text:
			if text := string(child.Literal); text != "" {
				inlines = append(inlines, textNode(text, marks))
			}
		case bf.Code:
			if text := string(child.Literal); text != "" {
				inlines = append(inlines, textNode(text, appendMark(codeCompatibleMarks(marks), adf.MarkCode, nil)))
			}
		case bf.Softbreak, bf.Hardbreak:
			inlines = append(inlines, &adf.Node{NodeType: adf.InlineNodeHardBreak})
		case bf.Strong:
			inlines = append(inlines, toADFInlines(child, appendMark(marks, adf.MarkStrong, nil))...)
		case bf.Emph:
			inlines = append(inlines, toADFInlines(child, appendMark(marks, adf.MarkEm, nil))...)
		case bf.Del:
			inlines = append(inlines, toADFInlines(child, appendMark(marks, adf.MarkStrike, nil))...)
		case bf.Link:
			inlines = append(inlines, toADFInlines(child, appendMark(marks, adf.MarkLink, map[string]any{
				"href": string(child.Destination),
			}))...)
		case bf.Image:
			if text := string(child.Title); text != "" {
				inlines = append(inlines, textNode(text, marks))
			}
		default:
			inlines = append(inlines, toADFInlines(child, marks)...)
		}
	}

	return inlines
}

func textNode(text string, marks []adf.MarkNode) *adf.Node {
	if len(marks) == 0 {
		return &adf.Node{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: text}}
	}

	copiedMarks := make([]adf.MarkNode, len(marks))
	copy(copiedMarks, marks)

	return &adf.Node{NodeType: adf.ChildNodeText, NodeValue: adf.NodeValue{Text: text, Marks: copiedMarks}}
}

func appendMark(marks []adf.MarkNode, markType adf.NodeType, attrs any) []adf.MarkNode {
	next := make([]adf.MarkNode, 0, len(marks)+1)
	next = append(next, marks...)
	next = append(next, adf.MarkNode{MarkType: markType, Attributes: attrs})
	return next
}

func codeCompatibleMarks(marks []adf.MarkNode) []adf.MarkNode {
	compatible := make([]adf.MarkNode, 0, len(marks))
	for _, mark := range marks {
		if mark.MarkType == adf.MarkLink {
			compatible = append(compatible, mark)
		}
	}
	return compatible
}
