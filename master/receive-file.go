package master

import (
	"encoding/json"
	"fmt"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	curl_http "github.com/headend/share-module/curl-http"
	file_and_directory "github.com/headend/share-module/file-and-directory"
	share_model "github.com/headend/share-module/model"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"log"
)

func MasterReceiveFile(msg string, gatewayUrl string, client *socketio_client.Client) {
	log.Printf("recieve the file: %v\n", msg)
	var fileInfoToRecieve share_model.WorkerUpdateSignal
	err := json.Unmarshal([]byte(msg), &fileInfoToRecieve)
	if err != nil {
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
		client.Emit(socket_event.ThongBao, "file ok")
	} else {
		print("file not ok")
		msg := fmt.Sprint("|%v| # origin |%v|", md5String, fileInfoToRecieve.Md5)
		client.Emit(socket_event.ThongBao, msg)
	}
}
