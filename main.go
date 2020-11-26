package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zhouhui8915/go-socket.io-client"
	"log"
	"os"
	"os/exec"
	"agent-getway-service/model"
	"iptv-agentd/utils"
)


func Shellout(shell string,command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(shell, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func main() {

	opts := &socketio_client.Options{
		Transport: "websocket",
		Query:     make(map[string]string),
	}
	opts.Query["user"] = "user"
	opts.Query["pwd"] = "pass"
	uri := "http://127.0.0.1:8000/socket.io/"

	client, err := socketio_client.NewClient(uri, opts)
	if err != nil {
		log.Printf("NewClient error:%v\n", err)
		return
	}

	client.On("error", func() {
		log.Printf("on error\n")
	})
	client.On("connection", func(msg string) {
		log.Printf("Connected whith message: %v\n", msg)
	})
	client.On("file", func(msg string) {
		log.Printf("recieve the file: %v\n", msg)
		var fileInfoToRecieve model.WorkerUpdateSignal
		err := json.Unmarshal([]byte(msg), &fileInfoToRecieve)
		if err != nil{
			print(err)
		}
		fmt.Printf("\n\n json object:::: %#v", fileInfoToRecieve)
		url := "http://127.0.0.1:8000/" + fileInfoToRecieve.FilePath
		utils.DownloadFile(url, fileInfoToRecieve.FilePath)
	})
	client.On("ok", func() {
		log.Printf("good\n")
	})
	client.On("message", func(msg string) {
		log.Printf("on message:%v\n", msg)
	})
	client.On("disconnection", func() {
		log.Printf("Dsconnect from server\nGoodbye!")
	})

	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		command := string(data)
		app := "echo"

		arg0 := "-e"
		arg1 := "1"
		arg2 := command


		cmd := exec.Command(app, arg0, arg1, arg2)
		stdout, err := cmd.Output()

		if err != nil {
			print(err.Error())
			return
		}

		client.Emit("notice", string(stdout))
		log.Printf("send message:%v\n", string(stdout))
		// xu l
		// call goi sang master
	}
}
