package dag

import (
	"container/list"
	"fmt"

	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"golang.org/x/exp/slices"
)

// DAG is a simple implementation of a directed acyclic graph.
type DAG struct {
	NameMap map[string]*Node
}

func NewDAG() *DAG {
	return &DAG{
		NameMap: make(map[string]*Node),
	}
}

type Node struct {
	Name     string
	Present  bool
	Parents  map[string]*Node
	Children map[string]*Node
}

func (d *DAG) Add(name string, dependants []string) (*Node, error) {
	n := d.getNode(name)
	n.Present = true

	dependantMap := make(map[string]bool)
	for _, dependant := range dependants {
		childrens := d.GetDeepChildren(name)
		ok := slices.Contains(childrens, dependant)
		if ok {
			return nil, fmt.Errorf("encountered circular dependency between %q and %q", name, dependant)
		}

		dependantMap[dependant] = true
	}

	for _, parent := range n.Parents {
		ok := dependantMap[parent.Name]
		if ok {
			delete(dependantMap, parent.Name)
		} else {
			d.removeChild(parent, name)
			delete(n.Parents, parent.Name)
		}
	}

	for newParent := range dependantMap {
		n.Parents[newParent] = d.addChild(newParent, n)
	}

	return n, nil
}

func (d *DAG) Delete(name string) {
	n := d.getNode(name)
	n.Present = false
	d.deleteBranch(n)
}

// GetDeepChildren will go down the DAG and get all children in the subtree
func (d *DAG) GetDeepChildren(name string) []string {
	children := make([]string, 0)

	n, ok := d.NameMap[name]
	if !ok {
		return []string{}
	}

	visited := make(map[string]*Node)
	// queue of the nodes to visit
	queue := list.New()
	queue.PushBack(n)
	// add the root node to the map of the visited nodes
	visited[n.Name] = n

	for queue.Len() > 0 {
		qnode := queue.Front()
		// iterate through all of its neighbors
		// mark the visited nodes; enqueue the non-visited
		for child, node := range qnode.Value.(*Node).Children {
			if _, ok := visited[child]; !ok {
				children = append(children, child)
				visited[child] = node
				queue.PushBack(node)
			}
		}
		queue.Remove(qnode)
	}

	return children
}

func (d *DAG) TopologicalSort() []string {
	visited := make(map[string]bool)
	stack := make([]string, 0)

	for _, node := range d.NameMap {
		if visited[node.Name] {
			continue
		}
		stack = d.topologicalSortHelper(node, visited, stack)
	}

	arrayutil.Reverse(stack)
	return stack
}

// GetChildren only returns the immediate children
func (d *DAG) GetChildren(name string) []string {
	n, ok := d.NameMap[name]
	if !ok {
		return []string{}
	}
	children := make([]string, 0)
	for childName, childNode := range n.Children {
		if !childNode.Present {
			continue
		}
		children = append(children, childName)
	}
	return children
}

// GetParents only returns the immediate parents
func (d *DAG) GetParents(name string) []string {
	n := d.getNode(name)
	parents := make([]string, 0)
	for _, parent := range n.Parents {
		if !parent.Present {
			continue
		}
		parents = append(parents, parent.Name)
	}
	return parents
}

func (d *DAG) Has(name string) bool {
	_, ok := d.NameMap[name]
	return ok
}

func (d *DAG) addChild(name string, child *Node) *Node {
	n := d.getNode(name)
	n.Children[child.Name] = child
	return n
}

func (d *DAG) removeChild(node *Node, childName string) {
	delete(node.Children, childName)
}

func (d *DAG) getNode(name string) *Node {
	n, ok := d.NameMap[name]
	if !ok {
		n = &Node{
			Name:     name,
			Parents:  make(map[string]*Node),
			Children: make(map[string]*Node),
		}
		d.NameMap[name] = n
	}
	return n
}

func (d *DAG) deleteBranch(n *Node) {
	if n.Present || len(n.Children) > 0 {
		return
	}

	for _, parent := range n.Parents {
		d.removeChild(parent, n.Name)
		d.deleteBranch(parent)
	}

	delete(d.NameMap, n.Name)
}

func (d *DAG) topologicalSortHelper(n *Node, visited map[string]bool, stack []string) []string {
	visited[n.Name] = true

	for _, childNode := range n.Children {
		if visited[childNode.Name] {
			continue
		}
		stack = d.topologicalSortHelper(childNode, visited, stack)
	}

	stack = append(stack, n.Name)

	return stack
}
