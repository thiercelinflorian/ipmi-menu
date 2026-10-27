package main

import "github.com/thiercelinflorian/ipmi-menu/cmd"

// version est renseignée à la compilation (-ldflags "-X main.version=…").
var version = "dev"

func main() { cmd.Execute(version) }
