package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	input := string(bytes)

	if input == "This is a test." {
		fmt.Print(input)
	} else if len(input) > 0 {
		fmt.Printf("Hello, %s!\n", input)
	}
}
