package main

import (
	"filesys/router"
)

func main() {
	r := router.InitRouter()
	r.Run(":8080")
}
