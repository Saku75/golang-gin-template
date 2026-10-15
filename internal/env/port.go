package env

import "strconv"

const (
	portKey     string = "PORT"
	portDefault int    = 8080
)

func Port() int {
	return lookup(portKey, portDefault, strconv.Atoi)
}
