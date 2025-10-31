package main

import (
	"videocall/cmd"

	_ "go.uber.org/automaxprocs"
)

func main() {
	cmd.Execute()
}
