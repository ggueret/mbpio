package main

import (
	"fmt"
	"encoding/json"
)

func PrintMap(class interface{}) {
	out, err := json.MarshalIndent(class, "", "  ")
	if err == nil {
		fmt.Print(string(out))
	}
}