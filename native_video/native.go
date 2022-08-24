package native_video

import (
	"encoding/json"
	"errors"
	"html"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/mxmCherry/openrtb/v15/native1/response"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/file_uploader"
	"github.com/prebid/prebid-server/filedownloader"
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
		glog.Error("Erro while unescaping Adm in bid", err.Error())
		return "", err
	}
	err = json.Unmarshal([]byte(unescaped), &navtiveResponse)
	if err != nil {
		glog.Error("Error while Unmarshalling native response", err.Error())
		return "", err
	}

	assestUrls := []string{}

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
			glog.Info("Will Support Title object soon")
		} else if asset.Video != nil {
			glog.Info("Found Video Asset in request")
			filePath, err := GetVideoFilePathFromVAST(asset.Video.VASTTag)
			if err != nil {
				glog.Error("Error while getting filepath from VAST", err.Error())
				return "", err
			}
			obj.FilePath = filePath
		} else if asset.Img != nil {
			obj.FilePath = asset.Img.URL
			obj.Subtype = BackgroundImage
		}
		assestUrls = append(assestUrls, obj.FilePath)
		path := strings.Split(obj.FilePath, "/")
		obj.FilePath = "/tmp/assets/" + path[len(path)-1]
		objectArray = append(objectArray, obj)
	}

	errs := filedownloader.DownloadMultipleFiles(assestUrls)
	if len(errs) != 0 {
		return "", errors.New("Error downloading assets")
	}

	num, err := strconv.Atoi(reqId)
	if err != nil {
		glog.Error("error while converting to number", err.Error())
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
	glog.Info("Going to store merged mp4 at ", mediaPath)

	currentTime := time.Now()
	Merge(AdTemplateMap[strconv.Itoa(int(num/10))], mediaPath, objectArray...)
	glog.Info("Time Taken for Merge Process :: ", time.Since(currentTime).Seconds())

	uploadCurrentTime := time.Now()
	uploadResponse, err := file_uploader.UploadAsset(mediaPath, uuid.String())
	if err != nil {
		glog.Error("Error while uploading merged video", err.Error())
		return "", err
	}
	glog.Info("Time Taken for Upload Process :: ", time.Since(uploadCurrentTime).Seconds())
	glog.Info("TIme taken for total process :: ", time.Since(currentTime).Seconds())

	vast := generateVASTXml("25", uploadResponse["url"])
	glog.Info("Final Vast Formed :: ", vast)

	err = filedownloader.RemoveAssets()
	if err != nil {
		glog.Error("Failed to clean up downloaded assets")
	}

	return vast, nil
}
