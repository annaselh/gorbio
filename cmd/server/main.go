package main

import (
	"log"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/inventory"
	"github.com/annaselh/gorbio/modules/sales"
)

func main() {
	app := core.New()

	app.Register(sales.New())
	app.Register(base.New())
	app.Register(inventory.New())

	if err := app.Boot(); err != nil {
		log.Fatalf("boot failed: %v", err)
	}

	if err := app.Serve(":8080"); err != nil {
		log.Fatal(err)
	}

}
