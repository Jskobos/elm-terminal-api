package main

import (
	"github.com/joho/godotenv"
)

func readEnv() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}	
}

func main() {
	readEnv()
	a := App{}
	a.Initialize()
	a.Run()
}