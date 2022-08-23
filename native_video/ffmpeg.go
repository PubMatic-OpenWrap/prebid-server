package native_video

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Subtype int

const (
	Main Subtype = iota
	Background
	Audio
	Image
	Title
)

type Object struct {
	ID       int
	FilePath string
	Duration int
	Price    float32
	Subtype  Subtype
}

/*
Template 1
-----------
*/
const template1Command string = `-y -i {{.BackgroundVideo}}
-i {{.MainVideo}}
-i {{.Audio1}}
-i {{.Audio2}}
-filter_complex [1:v]colorkey=0x14db04:0.3:0.2[ckout];[0:v][ckout]overlay[out];[2:a][3:a]amerge=inputs=2[a]
-map [out] -map [a]
-acodec ac3_fixed
-vcodec libx264
{{.OutputFile}}
`

// template used for replacing values in go template
type templateAttr struct {
	BackgroundVideo string
	MainVideo       string
	Audio1          string
	Audio2          string
	OutputFile      string
}

var template1 *template.Template
var err error

func init() {
	template1, err = template.New("template_1").Parse(template1Command)
	if err != nil {
		panic("Error in initializing template 1")
	}
}

func Merge(templateId int, impId string, objects ...Object) {

	mergedFilePath := filepath.Join("/tmp", impId)
	_ = os.MkdirAll(mergedFilePath, os.ModePerm)
	tattrbs := templateAttr{
		OutputFile: mergedFilePath + "/test.mp4",
	}
	for _, object := range objects {
		switch object.Subtype {
		case Background:
			tattrbs.BackgroundVideo = object.FilePath
		case Main:
			tattrbs.MainVideo = object.FilePath

		case Audio:
			if object.ID == 1 {
				tattrbs.Audio1 = object.FilePath
			}
			if object.ID == 2 {
				tattrbs.Audio2 = object.FilePath
			}
		default:
			panic("Invalid subtype")
		}
	}

	var processedTemplate1 bytes.Buffer
	template1.Execute(&processedTemplate1, tattrbs)
	fmt.Println(processedTemplate1.String())

	str := strings.ReplaceAll(processedTemplate1.String(), "\n", " ")
	var sarr []string
	for _, str := range strings.Split(str, " ") {
		if str != "" {
			sarr = append(sarr, str)
		}
	}

	cmd := exec.Command("ffmpeg", sarr...)
	fmt.Printf("Command to Execute : %v\n", cmd.String())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	// using stderr instead of out
	// Though there is no error
	// output is stored inside stderr
	fmt.Println(stderr.String())
	fmt.Println(out.String())

	/*
		ffmpeg -y -i http://localhost/hack22/background_AdobeExpress.mp4 -i http://localhost/hack22/moving_car_AdobeExpress.mp4 -i http://localhost/hack22/moving_car.mp3 -i http://localhost/hack22/bmw_audio_ad.mp3 -filter_complex '[1:v]colorkey=0x14db04:0.3:0.2[ckout];[0:v][ckout]overlay[out];[2:a][3:a]amerge=inputs=2[a]' -map '[out]' -map "[a]" -acodec ac3_fixed -vcodec libx264  /tmp/out2_audio_http.mp4
	*/
}
