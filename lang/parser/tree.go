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

	point := toPoint(location)

	node := pt.tree.RootNode().NamedDescendantForPointRange(point, point)
	if node == nil {
		return model.Symbol{}, false
	}

	for node != nil {
		if !pt.adapter.IsFunctionDecl(node) {
			node = node.Parent()

			continue
		}

		nameNode := node.ChildByFieldName("name")
		if nameNode == nil {
			return model.Symbol{}, false
		}

		return pt.symbolFromNode(nameNode), true
	}

	return model.Symbol{}, false
}

// FindFunctionByName finds all occurrences of a function with the specified name.
func (pt *ParseTree) FindFunctionByName(name string) []model.Symbol {
	results := []model.Symbol{}
	if pt.tree == nil {
		return results
	}

	var find func(*ts.Node)

	find = func(node *ts.Node) {
		if pt.adapter.IsFunctionDecl(node) {
			nameNode := node.ChildByFieldName("name")

			if nameNode != nil && nameNode.Utf8Text(pt.source) == name {
				results = append(results, pt.symbolFromNode(nameNode))
			}
		}

		for childIndex := range node.NamedChildCount() {
			find(node.NamedChild(childIndex))
		}
	}

	find(pt.tree.RootNode())

	return results
}

// FindFunctionCalls in a function body (funcDef is the function name).
func (pt *ParseTree) FindFunctionCalls(funcDef model.Location) []model.Symbol {
	point := toPoint(funcDef)
	node := pt.tree.RootNode().DescendantForPointRange(point, point)

	for node != nil && !pt.adapter.IsFunctionDecl(node) {
		node = node.Parent()
	}

	if node == nil {
		return nil
	}

	body := node.ChildByFieldName("body")
	if body == nil {
		return nil
	}

	var (
		result []model.Symbol
		find   func(*ts.Node)
	)

	find = func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			fn := pt.adapter.GetFuncCall(n.ChildByFieldName("function"))
			if fn != nil {
				result = append(result, pt.symbolFromNode(fn))
			}
		}

		for i := range n.NamedChildCount() {
			find(n.NamedChild(i))
		}
	}

	find(body)

	return result
}

// Close the tree.
func (pt *ParseTree) Close() {
	if pt.tree != nil {
		pt.tree.Close()
	}
}

func toPos(p ts.Point) model.Position {
	return model.Position{
		Line:      p.Row,
		Character: p.Column,
	}
}

func toPoint(l model.Location) ts.Point {
	return ts.Point{
		Row:    l.Range.Start.Line,
		Column: l.Range.Start.Character,
	}
}

func (pt *ParseTree) symbolFromNode(node *ts.Node) model.Symbol {
	name := node.Utf8Text(pt.source)
	location := model.Location{
		URI: pt.uri,
		Range: model.Range{
			Start: toPos(node.StartPosition()),
			End:   toPos(node.EndPosition()),
		},
	}

	return model.Symbol{
		ID:       model.SymbolID(location, name),
		Name:     name,
		Location: location,
	}
}
