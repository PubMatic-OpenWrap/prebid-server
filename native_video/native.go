package native_video

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/gofrs/uuid"
	"github.com/mxmCherry/openrtb/v15/native1/response"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/file_uploader"
)

func GetVideoFilePathFromVAST(vastBody string) (string, error) {

	vastBody = html.UnescapeString(vastBody)
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

type NativeResp struct {
	Native response.Response `json:"native"`
}

func ParseNativeVideoAdm(reqId string, bid *openrtb2.Bid, cacheId string) (string, error) {

	var navtiveResponse NativeResp
	unescaped, err := url.QueryUnescape(bid.AdM)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(unescaped), &navtiveResponse)
	if err != nil {
		return "", err
	}

	var objectArray []Object
	for _, asset := range navtiveResponse.Native.Assets {
		var assetExt map[string]interface{}
		err := json.Unmarshal(asset.Ext, &assetExt)
		if err != nil {
			return "", err
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
				return "", err
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
		return "", err
	}

	uuid, err := uuid.NewV1()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll("/tmp", os.ModePerm); err != nil {
		return "", err
	}
	mediaPath := "/tmp/" + uuid.String() + ".mp4"
	Merge(AdTemplateMap[strconv.Itoa(int(num/10))], mediaPath, objectArray...)

	uploadResponse, err := file_uploader.UploadAsset(mediaPath, uuid.String())
	if err != nil {
		return "", nil
	}
	vast := generateVASTXml("25", uploadResponse["url"])

	return vast, nil
}
