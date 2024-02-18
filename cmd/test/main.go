package main

import (
	"encoding/json"
	"fmt"

	"human.io/pulse/pkg/logic"
)

func toJSON(v any) string {
	js, _ := json.Marshal(v)
	return string(js)
}

func main() {
	scr := logic.Script{
		Functions: map[string]*logic.Function{
			"int.constant": {},
			"int.add":      {},
			"print":        {Execute: func() {}},
		},
		Nodes: map[int]*logic.Node{
			0: {Function: "int.add", X: 100, Y: 100, Inputs: map[string]logic.NodeInput{"a": {Node: 3, Output: "value"}, "b": {Node: 4, Output: "value"}}},
			1: {Function: "int.add", X: 100, Y: 300, Inputs: map[string]logic.NodeInput{"a": {Node: 0, Output: "result"}, "b": {Node: 4, Output: "value"}}},
			2: {Function: "print", X: 300, Y: 200, Inputs: map[string]logic.NodeInput{"value": {Node: 1, Output: "result"}}},
			3: {Function: "int.constant", X: -100, Y: 50, Constants: map[string]any{"value": 2}},
			4: {Function: "int.constant", X: -100, Y: 250, Constants: map[string]any{"value": 3}},
		},
	}
	nodeParents, err := scr.NodeParentScan(2)
	if err != nil {
		panic(err)
	}
	fmt.Println("nodeParents", toJSON(nodeParents))
	order, err := logic.ResolveOrder(nodeParents)
	if err != nil {
		panic(err)
	}
	fmt.Println("order", toJSON(order))
	// output, err := scr.NodeExecute(3, nil)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("output", toJSON(output))
	// fmt.Println("RUN")
	// nodeInput := map[int]map[string]any{}
	// for _, id := range order {
	// 	nodeInput[id] = map[string]any{}
	// }
	// for _, id := range order {
	// 	output, err := scr.NodeExecute(id, nodeInput[id])
	// 	log.Println("run", id, nodeInput[id], "->", output, err)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	for k, v := range output {
	// 		for _, connection := range scr.Connections {
	// 			if connection.FromNode == id && connection.FromOutput == k {
	// 				nodeInput[connection.ToNode][connection.ToInput] = v
	// 			}
	// 		}
	// 	}
	// }
}
