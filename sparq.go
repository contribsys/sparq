package sparq

import "fmt"

const (
	Name    = "Sparq"
	Version = "0.0.1"
)

var (
	UserAgent = fmt.Sprintf("%s v%s", Name, Version)
)
