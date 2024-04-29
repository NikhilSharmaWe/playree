package main

import (
	"log"
	"os"

	"github.com/NikhilSharmaWe/playree/playree/app"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("vars.env"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	application, err := app.NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	e := application.Router()

	log.Fatal(e.Start(os.Getenv("ADDR")))
}
