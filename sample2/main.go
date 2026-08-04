package main

import "github.com/kainkent69/fasttp/sample2/app"

func main() {
	r := app.SetupRouter()
	r.Run(":8080")
}
