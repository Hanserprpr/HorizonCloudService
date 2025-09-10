package main

import (
	"fmt"
	"system-service/internal/repository"
)

func main() {
	fmt.Println("Testing imports...")
	fmt.Printf("SystemSetting type: %T\n", repository.SystemSetting{})
	fmt.Println("Import test successful!")
}