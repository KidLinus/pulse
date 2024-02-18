package logic

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type Function struct {
	Input   any
	Output  any
	Execute func() error
	Read    func(*ReadContext) error
}

type Node struct {
	Function  string
	Constants map[string]any
	Inputs    map[string]NodeInput
	X, Y      int
}

type NodeInput struct {
	Node   int
	Output string
}

type Script struct {
	Functions map[string]*Function
	Nodes     map[int]*Node
}

type ReadContext struct {
	context.Context
	Function *Function
	Node     *Node
	Output   any
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func (ctx *ReadContext) InputBind(target any) error {
	t := reflect.TypeOf(target).Elem()
	rv := reflect.ValueOf(target).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		if tag, ok := field.Tag.Lookup("logic"); ok {
			name = tag
		}
		if v, ok := ctx.Node.Constants[name]; ok {
			fmt.Println("constant found", name, v)
			if !field.Type.AssignableTo(reflect.TypeOf(v)) {
				fmt.Println("	constant not assignable", field.Type, "!=", reflect.TypeOf(v))
			}
			rv.Field(i).Set(reflect.ValueOf(v))
			continue
		}
		if v, ok := ctx.Node.Inputs[name]; ok {
			fmt.Println("input found", name, v)
			continue
		}
	}
	fmt.Printf("Input after bind -> %#v\n", target)
	return nil
}

func (ctx *ReadContext) OutputSet(v any) error {
	ctx.Output = v
	return nil
}

func (s *Script) NodeParentScan(nodeID int) (map[int][]int, error) {
	nodeParents := map[int][]int{}
	scanQueue, discovered := []int{nodeID}, map[int]struct{}{nodeID: {}}
	for {
		if len(scanQueue) == 0 {
			break
		}
		id := scanQueue[0]
		scanQueue = scanQueue[1:]
		node, ok := s.Nodes[id]
		if !ok {
			return nil, errors.New("node not found")
		}
		parents, parentsM := []int{}, map[int]struct{}{}
		for _, input := range node.Inputs {
			if _, ok := parentsM[input.Node]; ok {
				continue
			}
			parents, parentsM[input.Node] = append(parents, input.Node), struct{}{}
			if _, ok := discovered[input.Node]; !ok {
				node, ok := s.Nodes[input.Node]
				if !ok {
					return nil, errors.New("node not found")
				}
				function, ok := s.Functions[node.Function]
				if !ok {
					return nil, errors.New("function not found")
				}
				discovered[input.Node] = struct{}{}
				if function.Execute == nil {
					scanQueue = append(scanQueue, input.Node)
					continue
				}
				nodeParents[input.Node] = []int{}
			}
		}
		nodeParents[id] = parents
	}
	return nodeParents, nil
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

func sliceIncludesMap[V comparable, A any](arr []V, vals map[V]A) bool {
	for _, v := range arr {
		if _, ok := vals[v]; !ok {
			return false
		}
	}
	return true
}

func NewFunctionPure[IN, OUT any](fn func(IN, *OUT) error) (*Function, error) {
	handler := func(ctx *ReadContext) error {
		input := new(IN)
		if err := ctx.InputBind(input); err != nil {
			return err
		}
		if err := validate.StructCtx(ctx, input); err != nil {
			return err
		}
		output := new(OUT)
		if err := fn(*input, output); err != nil {
			return err
		}
		return ctx.OutputSet(*output)
	}
	return &Function{
		Input:  *new(IN),
		Output: *new(IN),
		Read:   handler,
	}, nil
}
