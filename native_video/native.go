package native_video

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/mxmCherry/openrtb/v15/native1/response"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
)

func GetVideoFilePathFromVAST(vastBody string) (string, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromString(vastBody)
	if err != nil {
		return "", err
	}
	mediaFile := doc.FindElement("VAST/Ad/InLine/Creatives/Creative/Linear/MediaFiles/MediaFile")

	return strings.TrimSpace(mediaFile.Text()), nil
}

func GetSubType(subtype string) Subtype {

	switch subtype {
	case "main":
		return Main
	case "background":
		return BackgroundVideo
	case "audio":
		return Audio
	case "image":
		return BackgroundImage
	case "title":
		return Title
	}

	return 0

}

func ParseNativeVideoAdm(reqId string, bid *openrtb2.Bid, cacheId string) error {

	var navtiveResponse response.Response
	err := json.Unmarshal([]byte(bid.AdM), &navtiveResponse)
	if err != nil {
		return err
	}

	var objectArray []Object
	for _, asset := range navtiveResponse.Assets {

		var assetExt map[string]interface{}
		err := json.Unmarshal(asset.Ext, &assetExt)
		if err != nil {
			return err
		}

		obj := Object{
			ID:      int(*asset.ID),
			Subtype: GetSubType(assetExt["subtype"].(string)),
		}

		if asset.Title != nil {
			fmt.Println("Title Object")
		} else if asset.Video != nil {
			filePath, err := GetVideoFilePathFromVAST(asset.Video.VASTTag)
			if err != nil {
				return err
			}
			obj.FilePath = filePath
		} else if asset.Img != nil {
			fmt.Println("image Object")
		}
		objectArray = append(objectArray, obj)
	}

	num, err := strconv.Atoi(reqId)
	if err != nil {
		fmt.Println("error while converting to number")
		return err
	}
	Merge(AdTemplateMap[strconv.Itoa(int(num/10))], bid.ImpID, objectArray...)

	vast := generateVASTXml("25", "https://tech-stack-mgmt.pubmatic.com/owtools/hackathon2k22/owtools/api/getbid?reqid=11")

	w := new(bytes.Buffer)
	enc := xml.NewEncoder(w)
	enc.Encode(vast)
	bid.AdM = w.String()
	return nil
}
