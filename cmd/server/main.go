package main

import (
	"fmt"
	"os"

	"github.com/amsokol/go-grpc-http-rest-microservice-tutorial/pkg/cmd"
)

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
