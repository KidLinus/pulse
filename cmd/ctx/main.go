package main

import (
	"context"
	"encoding/json"
	"fmt"

	"human.io/pulse/pkg/logic"
)

func toJSON(v any) string {
	js, _ := json.Marshal(v)
	return string(js)
}

func IntAdd(in struct {
	A int `validate:"min=10"`
	B int
}, out *struct{ Result int }) error {
	out.Result = in.A + in.B
	return nil
}

func IntGt(in struct{ A, B int }, out *struct{ Result bool }) error {
	out.Result = in.A > in.B
	return nil
}

var scr = logic.Script{
	Functions: map[string]*logic.Function{},
	Nodes: map[int]*logic.Node{
		0: {Function: "int.add", X: 100, Y: 100, Constants: map[string]any{"A": 123}, Inputs: map[string]logic.NodeInput{"B": {Node: 4, Output: "value"}}},
		1: {Function: "int.add", X: 100, Y: 300, Inputs: map[string]logic.NodeInput{"a": {Node: 0, Output: "result"}, "b": {Node: 4, Output: "value"}}},
		2: {Function: "print", X: 300, Y: 200, Inputs: map[string]logic.NodeInput{"value": {Node: 1, Output: "result"}}},
		3: {Function: "int.constant", X: -100, Y: 50, Constants: map[string]any{"value": 2}},
		4: {Function: "int.constant", X: -100, Y: 250, Constants: map[string]any{"value": 3}},
	},
}

func main() {
	handler, err := logic.NewFunctionPure(IntAdd)
	if err != nil {
		panic(err)
	}
	scr.Functions["int.add"] = handler
	ctx := &logic.ReadContext{
		Context:  context.Background(),
		Function: handler,
		Node:     scr.Nodes[0],
	}
	if err := handler.Read(ctx); err != nil {
		panic(err)
	}
	fmt.Println("res", toJSON(ctx.Output))
}
