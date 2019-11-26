package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
	"github.com/labstack/gommon/log"
	"github.com/sanity-io/litter"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"os"
	"path/filepath"
)

var BUCKET string

func main() {
	BUCKET = os.Getenv(`IMAGE_BUCKET`)
	lambda.Start(serveFunc)
}

func serveFunc(request events.APIGatewayProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	log.Info(`start`)
	sess, err := session.NewSession()
	if err != nil {
		log.Error(err)
		return
	}

	s3Sdk :=  s3.New(sess)
	path := request.Path

	obj  , err := s3Sdk.GetObject(&s3.GetObjectInput{
		Bucket:                     &BUCKET,
		Key:                        &path,
	})
	if err != nil {
		log.Error(err)
		return
	}

	dst := []byte{}
	mimeType := mime.TypeByExtension(filepath.Ext(path))

	params := getParams(request.QueryStringParameters)
	if params.HasOptions() {
		dst  , mimeType , err = imagProcess(obj.Body , params)
		if err != nil {
			return
		}

	} else {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(obj.Body)
		if err != nil {
			return
		}
		dst = buf.Bytes()

	}

	sEnc := base64.StdEncoding.EncodeToString(dst)

	resp = events.APIGatewayProxyResponse{
		StatusCode:        200,
		Headers:           map[string]string{
			`Content-Type` : mimeType,
		},
		//MultiValueHeaders: nil,
		Body:              sEnc,
		IsBase64Encoded:   true,
	}

	return resp , nil
}

func imagProcess(r io.ReadCloser  , params Params) (encoded []byte , mimeType string , err error) {

	img, t, err := image.Decode(r)
	if err != nil {
		log.Error(err)
		return nil ,"" ,  err
	}

	decodeType  , mimeType := typeToDecodeAndMime(t)

	//rectange of image
	rctSrc := img.Bounds()

	litter.Dump(params)

	w ,  h := func() (int  ,int){
		if params.Width == nil && params.Height == nil  {
			return rctSrc.Dx() , rctSrc.Dy()
		}

		if params.Width != nil && params.Height != nil  {
			return *params.Width , *params.Height
		}

		if params.Width != nil {
			ratio := float64(rctSrc.Dy()) / float64(rctSrc.Dx())
			return *params.Width , int(float64(*params.Width) * ratio)
		}

		ratio := float64(rctSrc.Dx()) / float64(rctSrc.Dy())
		return int(float64(*params.Height) * ratio)  , *params.Height

	}()

	dst := new(bytes.Buffer)
	imgDst := imaging.Resize(img, w, h, imaging.Lanczos)

	//encode resized image
	/*
	dst := new(bytes.Buffer)

	 */
	err = imaging.Encode(dst , imgDst , decodeType)
	if err != nil {
		log.Error(err)
		return nil, "",  err
	}
	return dst.Bytes() , mimeType  , nil

}

func typeToDecodeAndMime(t string) (imaging.Format , string) {

	switch t {
	case "jpeg":
			return imaging.JPEG , `image/jpeg`
	case "gif":
			return imaging.GIF , `image/gif`
	case "png":
			return imaging.PNG  ,`image/png`
	case "tiff":
		return imaging.TIFF  ,`image/tiff`
	default:
		return imaging.JPEG , `image/jpeg`
	}
}


type Params struct {
	Width *int `json:"width,string,omitempty"`
	Height *int  `json:"height,string,omitempty"`
}

func (self *Params) HasOptions () bool {
	if self.Width != nil {
		return true
	}
	if self.Height != nil {
		return true
	}
	return false
}


func getParams(m map[string]string)  Params{
	p := Params{}
	if len(m)  ==0 {
		return p
	}

	j , err := json.Marshal(m)
	if err != nil {
		log.Error(err)
	}

	err = json.Unmarshal(j, &p)
	if err != nil {
		log.Error(err)
	}

	return p
}