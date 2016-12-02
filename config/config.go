package config

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type responder struct {
	Regexp    string
	Responses []string
}

type responders struct {
	Rs []responder
}

// PopulateResponders reads the toml file "responders.toml" and returns
// a slice of responders read from the config file
func PopulateResponders() []responder {
	var rs responders
	if _, err := toml.DecodeFile("responders.toml", &rs); err != nil {
		log.Fatalf("Error reading responders file: %v", err)
	}
	fmt.Println(rs)
	return rs.Rs
}
