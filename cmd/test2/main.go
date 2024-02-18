package main

import (
	"encoding/json"
	"fmt"
	"log"

	"human.io/pulse/pkg/logic"
)

func toJSON(v any) string {
	js, _ := json.Marshal(v)
	return string(js)
}

func main() {
	scr := logic.Script{
		Functions: map[string]*logic.Function{
			"int.constant": {
				Name:        "Int - Constant",
				Description: "A single int value",
				Inputs:      map[string]logic.Input{"value": {Type: "int"}},
				Output:      map[string]logic.Output{"value": {Type: "int"}},
				Pure:        true,
				Handler:     func(m map[string]any) (map[string]any, error) { return m, nil },
			},
			"int.add": {
				Name:        "Int - Summarize",
				Description: "Add multiple ints together",
				Inputs: map[string]logic.Input{
					"a": {Type: "int"},
					"b": {Type: "int"},
				},
				Output: map[string]logic.Output{
					"result": {Type: "int"},
				},
				Pure: true,
				Handler: func(m map[string]any) (map[string]any, error) {
					return map[string]any{"result": m["a"].(int) + m["b"].(int)}, nil
				},
			},
			"print": {
				Name:        "Print value",
				Description: "Print any value into the console",
				Inputs: map[string]logic.Input{
					"value": {Type: "any"},
				},
				Handler: func(m map[string]any) (map[string]any, error) {
					log.Println("PRINT", m["value"])
					return map[string]any{}, nil
				},
			},
		},
		Nodes: map[int]*logic.Node{
			0: {Function: "int.add", X: 100, Y: 100},
			1: {Function: "int.add", X: 100, Y: 300},
			2: {Function: "print", X: 300, Y: 200},
			3: {Function: "int.constant", X: -100, Y: 50, Constants: map[string]any{"value": 2}},
			4: {Function: "int.constant", X: -100, Y: 250, Constants: map[string]any{"value": 3}},
		},
		Connections: map[int]*logic.Connection{
			0: {FromNode: 3, FromOutput: "value", ToNode: 0, ToInput: "a"},
			1: {FromNode: 4, FromOutput: "value", ToNode: 0, ToInput: "b"},
			2: {FromNode: 0, FromOutput: "result", ToNode: 1, ToInput: "a"},
			3: {FromNode: 4, FromOutput: "value", ToNode: 1, ToInput: "b"},
			4: {FromNode: 1, FromOutput: "result", ToNode: 2, ToInput: "value"},
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
	output, err := scr.NodeExecute(3, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("output", toJSON(output))
	fmt.Println("RUN")
	nodeInput := map[int]map[string]any{}
	for _, id := range order {
		nodeInput[id] = map[string]any{}
	}
	for _, id := range order {
		output, err := scr.NodeExecute(id, nodeInput[id])
		log.Println("run", id, nodeInput[id], "->", output, err)
		if err != nil {
			panic(err)
		}
		for k, v := range output {
			for _, connection := range scr.Connections {
				if connection.FromNode == id && connection.FromOutput == k {
					nodeInput[connection.ToNode][connection.ToInput] = v
				}
			}
		}
	}
}
