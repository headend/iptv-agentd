package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/headend/share-module/curl-http"
	"github.com/headend/share-module/file-and-directory"
	share_model "github.com/headend/share-module/model"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/static-config"

	"github.com/zhouhui8915/go-socket.io-client"
	"log"
	"os"
	"os/exec"
)


func main() {

	// load config
	var conf  configuration.Conf
	conf.LoadConf()
	//log.Printf("%#v", conf)

	/*
	Xử lý thông tin kết nối
	Nếu thông tin không có trong config thì lấy từ static config
	*/
	var gwHost string
	if conf.AgentGateway.Gateway != "" {
		gwHost = conf.AgentGateway.Gateway
	} else {
		if conf.AgentGateway.Host != "" {
			gwHost = conf.AgentGateway.Host
		} else {
			gwHost = static_config.GatewayHost
		}
	}
	var gwPort uint16
	if conf.AgentGateway.Port != 0 {
		gwPort = conf.AgentGateway.Port
	} else {
		gwPort = static_config.GatewayPort
	}
	// make authen params
	opts := &socketio_client.Options{
		Transport: "websocket",
		Query:     make(map[string]string),
	}
	opts.Query["user"] = static_config.GatewayUser
	opts.Query["pwd"] = static_config.GatewayPassword
	var gatewayUrl string
	gatewayUrl = fmt.Sprintf("http://%s:%d/", gwHost, gwPort)
	//gatewayUrl = "http://" + gwHost + ":" + string(gwPort) + "/"
	println(gatewayUrl)
	var uri string
	uri = gatewayUrl+ "socket.io/"

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
		url := gatewayUrl + fileInfoToRecieve.FilePath
		curl_http.DownloadFile(url, fileInfoToRecieve.FilePath)
		md5String, err := file_and_directory.GetMd5FromFile(fileInfoToRecieve.FilePath)
		if err != nil {
			print(err)
		}
		// compare md5
		if md5String == fileInfoToRecieve.Md5 {
			print("file ok")
			client.Emit("notice", "file ok")
		} else {
			print("file not ok")
			msg := fmt.Sprint("|%v| # origin |%v|", md5String, fileInfoToRecieve.Md5)
			client.Emit("notice", msg)
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
