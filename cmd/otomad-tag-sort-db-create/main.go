package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/naari3/otomad-tag-sort/cmd/otomad-tag-sort-db-create/cmd"
)

func main() {
	cmd.Execute()
}
