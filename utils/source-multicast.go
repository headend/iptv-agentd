package utils

import (
	"encoding/json"
	"fmt"
	static_config "github.com/headend/share-module/configuration/static-config"
	"github.com/headend/share-module/model"
	"github.com/headend/share-module/shellout"
	"log"
)

func CheckSourceMulticast(sourceMulticast string) (err error, sourceStatusCode int64) {
	shell := fmt.Sprintf("%s/%s", static_config.BinaryPath, "ffprobe")
	cmdtorun := []string{"udp://"+sourceMulticast, "-v", "quiet", "-show_format", "-show_streams", "-print_format", "json"}
	err, _, stdOut, _ := shellout.RunExternalCmd(shell, cmdtorun, 30)

	if err != nil {
		//log.Println(err.Error())
		//log.Println(exitCode)
		//log.Println(stdOut)
		//log.Println(stdErr)
		return err, static_config.SourceNotOK
	}
	/*
	convert stdOut to json
	 */
	//log.Print(stdOut)
	var sourceInfo model.FfprobeResponse
	err = json.Unmarshal([]byte(stdOut), &sourceInfo)
	if err != nil {
		print(err)
		return err, static_config.SourceUnknow
	}

	/*
	check audio, video
	*/
	var  audio, video int
	if sourceInfo.Format.Filename == "" {

		return nil, static_config.SourceNotOK
	}
	for _,stream := range sourceInfo.Streams {
		switch stream.Codec_type {
		case "video":
			video = 1
		case "audio":
			audio = 1
		}
	}
	if video == 1 && audio ==1 {
		return nil, static_config.SourceOK
	}
	if video == 0 && audio == 1{
		return nil, static_config.SourceNoVideo
	}
	if video == 1 && audio == 0 {
		return nil, static_config.SourceNoAudio
	}

	log.Printf("json %#v", sourceInfo)
	err = fmt.Errorf("Unknow")
	return err, static_config.SourceNotOK
}

func Capture(sourceMulticast string, pathToSaveImageFile string) error {
	shell := fmt.Sprintf("%s/%s", static_config.BinaryPath, "ffprobe")
	commandToRun := []string{"-timeout","28","-i", sourceMulticast, "-v", "quiet","-r","1","-f","image2",pathToSaveImageFile,"-y"}
	err, exitCode, _, stdErr := shellout.RunExternalCmd(shell, commandToRun, 30)
	if err != nil{
		return err
	}
	if exitCode != 0 {
		err = fmt.Errorf("Meet error with exit code: %d and message: %s", exitCode, stdErr)
		return err
	}
	return nil
}
