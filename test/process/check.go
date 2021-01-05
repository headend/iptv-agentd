package main

import (
	"log"
	"os"
	"syscall"
)

func main()  {
	process, err := os.FindProcess(10519)
	if err != nil {
		log.Println(err)
		return
	}
	killErr := process.Signal(syscall.Signal(0))
	procExists := killErr == nil
	println("Exists: ", procExists)
}
