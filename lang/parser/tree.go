package parser

import (
	ts "github.com/tree-sitter/go-tree-sitter"
	"github.com/vulns-are-features-too/func-tracer/model"
)

// ParseTree is the parsed content of a file.
type ParseTree struct {
	adapter Adapter
	uri     string
	tree    *ts.Tree
	source  []byte
}

// FindFunction by location.
func (pt *ParseTree) FindFunction(location model.Location) (model.Symbol, bool) {
	if pt.tree == nil {
		return model.Symbol{}, false
	}

	point := ts.Point{
		Row:    location.Range.Start.Line,
		Column: location.Range.Start.Character,
	}

	node := pt.tree.RootNode().NamedDescendantForPointRange(point, point)
	if node == nil {
		return model.Symbol{}, false
	}

	for node != nil {
		if !pt.adapter.IsFunctionKind(node.Kind()) {
			node = node.Parent()

			continue
		}

		nameNode := node.ChildByFieldName("name")
		if nameNode == nil {
			return model.Symbol{}, false
		}

		name := nameNode.Utf8Text(pt.source)

		return pt.symbolFromNode(nameNode, name), true
	}

	return model.Symbol{}, false
}

// FindFunctionByName finds the 1st occurrence of a function with the specified name.
func (pt *ParseTree) FindFunctionByName(name string) (model.Symbol, bool) {
	if pt.tree == nil {
		return model.Symbol{}, false
	}

	var find func(*ts.Node) (model.Symbol, bool)

	find = func(node *ts.Node) (model.Symbol, bool) {
		if pt.adapter.IsFunctionKind(node.Kind()) {
			nameNode := node.ChildByFieldName("name")

			if nameNode != nil && nameNode.Utf8Text(pt.source) == name {
				return pt.symbolFromNode(nameNode, name), true
			}
		}

		for childIndex := range node.NamedChildCount() {
			child := node.NamedChild(childIndex)
			if symbol, ok := find(child); ok {
				return symbol, true
			}
		}

		return model.Symbol{}, false
	}

	return find(pt.tree.RootNode())
}

// Close the tree.
func (pt *ParseTree) Close() {
	if pt.tree != nil {
		pt.tree.Close()
	}
}

func (pt *ParseTree) symbolFromNode(
	node *ts.Node,
	name string,
) model.Symbol {
	start := node.StartPosition()
	end := node.EndPosition()

	location := model.Location{
		URI: pt.uri,
		Range: model.Range{
			Start: model.Position{
				Line:      start.Row,
				Character: start.Column,
			},
			End: model.Position{
				Line:      end.Row,
				Character: end.Column,
			},
		},
	}

	return model.Symbol{
		ID:       model.SymbolID(location, name),
		Name:     name,
		Location: location,
	}
}
