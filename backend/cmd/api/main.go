package main

import (
	"fmt"

	"github.com/seymourrisey/payflow-simulator/config"
)

func main() {
	fmt.Println("Starting PayFlow API...")
	config.Load()
	config.ConnectDB()
	defer config.CloseDB()
}
