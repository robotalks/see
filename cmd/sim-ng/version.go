package main

import "fmt"

// Version number
const Version = "0.1.0"

// VersionSuffix provides suffix info
var VersionSuffix = "-dev"

// PrintVersion prints version
func PrintVersion() {
	fmt.Println(Version + VersionSuffix)
}

type verCmd struct {
}

func (c *verCmd) Execute(args []string) error {
	PrintVersion()
	return nil
}
