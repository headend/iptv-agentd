package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/headend/share-module/curl-http"
	"github.com/headend/share-module/file-and-directory"
	share_model "github.com/headend/share-module/model"
	"github.com/zhouhui8915/go-socket.io-client"
	"log"
	"os"
	"os/exec"
)


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
		var fileInfoToRecieve share_model.WorkerUpdateSignal
		err := json.Unmarshal([]byte(msg), &fileInfoToRecieve)
		if err != nil{
			print(err)
		}
		fmt.Printf("\n\n json object:::: %#v", fileInfoToRecieve)
		url := "http://127.0.0.1:8000/" + fileInfoToRecieve.FilePath
		curl_http.DownloadFile(url, fileInfoToRecieve.FilePath)
		md5String, err := file_and_directory.GetMd5FromFile(fileInfoToRecieve.FilePath)
		if err != nil {
			print(err)
		}
		// compare md5
		if md5String == fileInfoToRecieve.Md5 {
			print("file ok")
		} else {
			print("file not ok")
			fmt.Printf("|%v| # origin |%v|", md5String, fileInfoToRecieve.Md5)
		}
	})

	client.On("message", func(msg string) {
		log.Printf("on message:%v\n", msg)
	})
	client.On("disconnection", func() {
		log.Printf("Disconnect from server\nGoodbye!")
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
		// xu lý
		// call goi sang master
	}
}
