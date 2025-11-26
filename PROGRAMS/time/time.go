package main

import (
	"fmt"
	"time"
)

func main() {
	// Return current time in a readable format
	now := time.Now()
	fmt.Printf("Current time: %s\n", now.Format("Monday, January 2, 2006 at 3:04:05 PM MST"))
}
