package main

import "github.com/kainkent69/fasttp/sample/app"

func main() {
	r := app.SetupRouter()
	r.Run(":8080")
}
