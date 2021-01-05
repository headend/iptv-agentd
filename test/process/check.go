package main

import (
	"log"
	"os"
	"syscall"
)

func main()  {
	var pid int
	pid = 10515
	procExists := isExistsProcess(pid)
	println("Exists: ", procExists)
}

func isExistsProcess(pid int) ( isExists bool) {
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Println(err)
		return false
	}
	killErr := process.Signal(syscall.Signal(0))
	procExists := killErr == nil
	return procExists
}
