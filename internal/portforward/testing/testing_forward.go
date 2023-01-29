package main

import (
	"fmt"
	"github.com/pytogo/pytogo/portforward"
	"time"
)

func main() {
	err := portforward.Forward("test", "nginx-service", 4000, 80, "/Users/sebastianziemann/.kube/config", 0, "")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		time.Sleep(1 * time.Minute)
	}
}
