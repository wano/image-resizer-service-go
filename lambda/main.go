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
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
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

	s3Sdk := s3.New(sess)
	path := request.Path

	obj, err := s3Sdk.GetObject(&s3.GetObjectInput{
		Bucket: &BUCKET,
		Key:    &path,
	})
	if err != nil {
		log.Error(err)
		return
	}

	original := new(bytes.Buffer)
	_, err = original.ReadFrom(obj.Body)
	if err != nil {
		log.Error(err)
		return
	}

	mimeAndDecodeType, err := getMimeAndDecodeType(original)
	if err != nil {
		log.Error(err)
		return
	}

	dst := []byte{}
	mimeType := mimeAndDecodeType.ContentType

	const LAMBDA_MAX_RESPONSE = 1024 * 1024 * 5

	params := getParams(request.QueryStringParameters)
	if params.HasOptions() {
		dst, mimeType, err = imagProcess(original, params, *mimeAndDecodeType)
		if err != nil {
			return
		}

	} else if *obj.ContentLength > LAMBDA_MAX_RESPONSE {
		forceResize := 1920
		dst, mimeType, err = imagProcess(original, Params{
			Width: &forceResize,
		}, *mimeAndDecodeType)
		if err != nil {
			return
		}

	} else {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(obj.Body)
		if err != nil {
			log.Error(err)
			return
		}
		dst = buf.Bytes()
	}

	sEnc := base64.StdEncoding.EncodeToString(dst)

	resp = events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			`Content-Type`: mimeType,
		},
		//MultiValueHeaders: nil,
		Body:            sEnc,
		IsBase64Encoded: true,
	}

	return resp, nil
}

func imagProcess(buf *bytes.Buffer, params Params, md MimeAndDecodeType) (encoded []byte, mimeType string, err error) {

	decodeOptions := []imaging.DecodeOption{}

	if params.AutoRotate {
		decodeOptions = append(decodeOptions, imaging.AutoOrientation(true))
	}

	img, err := imaging.Decode(buf, decodeOptions...)
	if err != nil {
		log.Error(err)
		return nil, "", err
	}

	rctSrc := img.Bounds()

	w, h := func() (int, int) {
		if params.Width == nil && params.Height == nil {
			return rctSrc.Dx(), rctSrc.Dy()
		}

		if params.Width != nil && params.Height != nil {
			return *params.Width, *params.Height
		}

		if params.Width != nil {
			ratio := float64(rctSrc.Dy()) / float64(rctSrc.Dx())
			return *params.Width, int(float64(*params.Width) * ratio)
		}

		ratio := float64(rctSrc.Dx()) / float64(rctSrc.Dy())
		return int(float64(*params.Height) * ratio), *params.Height

	}()

	dst := new(bytes.Buffer)
	imgDst := imaging.Resize(img, w, h, imaging.Lanczos)

	err = imaging.Encode(dst, imgDst, md.Format)
	if err != nil {
		log.Error(err)
		return nil, "", err
	}
	return dst.Bytes(), mimeType, nil

}

type Params struct {
	Width      *int `json:"width,string,omitempty"`
	Height     *int `json:"height,string,omitempty"`
	AutoRotate bool `json:"auto_rotate,string,omitempty"`
}

func (self *Params) HasOptions() bool {
	if self.Width != nil {
		return true
	}
	if self.Height != nil {
		return true
	}
	return false
}

func getParams(m map[string]string) Params {
	p := Params{}
	if len(m) == 0 {
		return p
	}

	j, err := json.Marshal(m)
	if err != nil {
		log.Error(err)
	}

	err = json.Unmarshal(j, &p)
	if err != nil {
		log.Error(err)
	}

	return p
}

func getMimeAndDecodeType(b *bytes.Buffer) (*MimeAndDecodeType, error) {

	// todo: be lightweight
	bb := b.Bytes()

	contentType := http.DetectContentType(bb)

	f := func() imaging.Format {
		switch contentType {
		case `image/jpeg`:
			return imaging.JPEG
		case "image/gif":
			return imaging.GIF
		case "image/png":
			return imaging.PNG
		case "image/tiff":
			return imaging.TIFF
		default:
			return imaging.JPEG
		}

	}()

	return &MimeAndDecodeType{
		ContentType: contentType,
		Format:      f,
	}, nil
}

type MimeAndDecodeType struct {
	ContentType string
	Format      imaging.Format
}
