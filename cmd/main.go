package main

import (
	"github.com/pro0o/yoo-chat/endpoints"
)

func main() {
	server := endpoints.NewAPIServer(":4747")
	server.Run()
}
