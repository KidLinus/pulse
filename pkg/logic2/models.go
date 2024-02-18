package logic

import (
	"errors"
	"fmt"
)

type Script struct {
	Functions   map[string]*Function
	Nodes       map[int]*Node
	Connections map[int]*Connection
}

type Function struct {
	Name        string
	Description string
	Inputs      map[string]Input
	Output      map[string]Output
	Handler     func(map[string]any) (map[string]any, error)
	Execute     func()
	State       func()
	Pure        bool
}

type Input struct {
	Type        string
	Validation  string
	Description string
}

type Output struct {
	Type        string
	Description string
}

type Node struct {
	Function  string
	Constants map[string]any
	X, Y      int
}

type Connection struct {
	FromNode   int
	FromOutput string
	ToNode     int
	ToInput    string
}

type Context struct {
	NodeExecutionOrder []int
	Nodes              map[int]*ContextNode
}

type ContextNode struct {
	Parents []int
}

func (s *Script) NodeParentScan(nodeID int) (map[int][]int, error) {
	parents := map[int][]int{}
	for _, connection := range s.Connections {
		if sliceIncludes(parents[connection.ToNode], connection.FromNode) {
			continue
		}
		parents[connection.ToNode] = append(parents[connection.ToNode], connection.FromNode)
	}
	nodes := map[int][]int{}
	queue, discovered := []int{nodeID}, map[int]struct{}{nodeID: {}}
	for {
		if len(queue) == 0 {
			break
		}
		id := queue[0]
		queue = queue[1:]
		nodes[id] = append([]int{}, parents[id]...)
		for _, parent := range parents[id] {
			if _, ok := discovered[parent]; ok {
				continue
			}
			discovered[parent] = struct{}{}
			node, ok := s.Nodes[parent]
			if !ok {
				return nil, errors.New("node not found")
			}
			function, ok := s.Functions[node.Function]
			if !ok {
				return nil, errors.New("function not found")
			}
			if !function.Pure {
				nodes[parent] = []int{}
				continue
			}
			queue = append(queue, parent)
		}
	}
	return nodes, nil
}

func sliceIncludes[V comparable](arr []V, val V) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func sliceIncludesMap[V comparable, A any](arr []V, vals map[V]A) bool {
	for _, v := range arr {
		if _, ok := vals[v]; !ok {
			return false
		}
	}
	return true
}

func (s *Script) NodeTrigger(nodeID int) error {
	nodeParents, err := s.NodeParentScan(nodeID)
	if err != nil {
		return err
	}
	order, done := []int{}, map[int]struct{}{}
	for {
		if len(nodeParents) == 0 {
			break
		}
		for id, parents := range nodeParents {
			if sliceIncludesMap(parents, done) {
				order, done[id] = append(order, id), struct{}{}
				delete(nodeParents, id)
				break
			}
		}
		return errors.New("unable to execute")
	}
	fmt.Println("order", order)
	//ctx := &Context{}
	return nil
}

func ResolveOrder(nodeParents map[int][]int) ([]int, error) {
	order, resolved := []int{}, map[int]struct{}{}
resolveLoop:
	for {
		if len(order) == len(nodeParents) {
			break
		}
		for id, parents := range nodeParents {
			if _, ok := resolved[id]; ok {
				continue
			}
			if sliceIncludesMap(parents, resolved) {
				order = append(order, id)
				resolved[id] = struct{}{}
				continue resolveLoop
			}
		}
		return nil, errors.New("not resolveable")
	}
	return order, nil
}

func (s *Script) NodeExecute(nodeID int, input map[string]any) (map[string]any, error) {
	node, ok := s.Nodes[nodeID]
	if !ok {
		return nil, errors.New("node not found")
	}
	function, ok := s.Functions[node.Function]
	if !ok {
		return nil, errors.New("function not found")
	}
	in := map[string]any{}
	for k, v := range input {
		in[k] = v
	}
	for k, v := range node.Constants {
		in[k] = v
	}
	output, err := function.Handler(in)
	if err != nil {
		return nil, err
	}
	return output, nil
}
