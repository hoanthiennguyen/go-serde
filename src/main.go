package main

import (
	"fmt"

	"github.com/hoanthiennguyen/go-serde/src/json"
)

func main() {
	fmt.Println(json.Serialize("hello"))
}
