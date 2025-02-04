package main

import (
	"os"

	"gabe565.com/linx-server/internal/config"
	"github.com/pelletier/go-toml/v2"
)

func main() {
	f, err := os.Create("config_example.toml")
	if err != nil {
		panic(err)
	}

	encoder := toml.NewEncoder(f)
	encoder.SetIndentTables(true)
	conf := config.New()
	if err := encoder.Encode(conf); err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}
}
