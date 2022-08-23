package native_video

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"strings"
)

/*
AdTemplate specifies the supported video native ad templates
*/
type AdTemplate string

/*
AdTemplate1
-----------
Overlays Main video with Background Video. It replaces (0x14db04) color. It also considers 2 audio streams
Following list represents the assets required
	1. Background - Video  (Mandatory)
	2. Main - Video (Mandatory)
	3. Audio1 (Mandatory)
	4. Audio 2 (Optional)
NOTE: Please ensure that Resolution and Duration of Main & Background Video is matching
*/
const AdTemplate3 AdTemplate = `-y -i {{.BackgroundVideo}}
-i {{.MainVideo}}
-i {{.Audio1}}
{{.Audio2}}
-filter_complex {{.FilterComplex}}
-map [out] -map [a]
-acodec ac3_fixed
-vcodec libx264
{{.OutputFile}}
`

const AdTemplate1 AdTemplate = `-y -i {{.BackgroundVideo}}
-i {{.MainVideo}}
-filter_complex [1:v]colorkey=0x14db04:0.3:0.2[ckout];[0:v][ckout]overlay[out]
-map [out]
-acodec ac3_fixed
-vcodec libx264
{{.OutputFile}}
`

/*
AdTemplate2
---------
Overlays Main video with Background Image. It replaces (0x14db04) color. It also considers 1 audio stream
Following list represents the assets required
	1. Background - Image  (Mandatory)
	2. Main - Video (Mandatory)
	3. Audio 1 (Mandatory)
*/
const AdTemplate2 AdTemplate = `-y -i {{.BackgroundImage}} 
-i {{.MainVideo}} 
-i {{.Audio1}} 
-filter_complex [1:v]chromakey=0x42FB00:0.1:0.2[ckout];[0:v][ckout]overlay[out]
-map [out]
-map 2
-shortest
{{.OutputFile}}
`

// NOTE: if you are adding new template. Please add its entry in below templateIDMap

var templateIDMap = map[AdTemplate]*template.Template{
	AdTemplate1: nil,
	AdTemplate2: nil,
}

type Subtype int

const (
	Main Subtype = iota
	BackgroundVideo
	Audio
	BackgroundImage
	Title
)

var AdTemplateMap = map[string]AdTemplate{
	"1": AdTemplate1,
	"2": AdTemplate2,
}

type Object struct {
	ID       int
	FilePath string
	Duration int
	Price    float32
	Subtype  Subtype
}

// template used for replacing values in go template
type templateAttr struct {
	BackgroundVideo string
	BackgroundImage string
	MainVideo       string
	Audio1          string
	Audio2          string
	OutputFile      string
	FilterComplex   string
}

func init() {
	for tmpl := range templateIDMap {
		adTemplateCompiled, err := template.New(string(tmpl)).Parse(string(tmpl))
		if err != nil {
			panic(fmt.Sprintf("Error in initializing template %v", tmpl))
		}
		templateIDMap[tmpl] = adTemplateCompiled
	}
}

/*
Merge will compile multiple viodes, audios, images in one single mp4 file
based on adTemplate. mp4 file will be generated at mergedFilePath
*/
func Merge(adTemplate AdTemplate, mergedFilePath string, objects ...Object) {
	tattrbs := templateAttr{
		OutputFile: mergedFilePath,
	}
	audioCount := 0
	for _, object := range objects {
		switch object.Subtype {
		case BackgroundVideo:
			tattrbs.BackgroundVideo = object.FilePath
		case Main:
			tattrbs.MainVideo = object.FilePath
		case Audio:
			audioCount++
			tattrbs.Audio1 = object.FilePath
			if audioCount > 1 {
				tattrbs.Audio2 = "-i " + object.FilePath
				if adTemplate == AdTemplate1 {
					tattrbs.FilterComplex = "[1:v]colorkey=0x14db04:0.3:0.2[ckout];[0:v][ckout]overlay[out];[2:a][3:a]amerge=inputs=2[a]"
				}
			}
			if adTemplate == AdTemplate1 {
				tattrbs.FilterComplex = "[1:v]colorkey=0x14db04:0.3:0.2[ckout];[0:v][ckout]overlay[out];[2:a]amerge=inputs=1[a]"
			}
		case BackgroundImage:
			tattrbs.BackgroundImage = object.FilePath
		default:
			panic("Invalid subtype")
		}
	}

	var processedTemplate1 bytes.Buffer
	templateIDMap[adTemplate].Execute(&processedTemplate1, tattrbs)
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
}
