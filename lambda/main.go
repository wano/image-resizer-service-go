package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
	"github.com/labstack/gommon/log"
	"github.com/mitchellh/mapstructure"
	"image"
	"io"
	"os"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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

	params := getParams(request.QueryStringParameters)
	if params.HasOptions() {
		dst  , err = imagProcess(obj.Body , params)
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
			`Content-Type` : `image/jpeg`,
		},
		//MultiValueHeaders: nil,
		Body:              sEnc,
		IsBase64Encoded:   true,
	}

	return resp , nil
}

func imagProcess(r io.ReadCloser  , params Params) ([]byte ,error) {

	img, t, err := image.Decode(r)
	if err != nil {
		log.Error(err)
		return nil , err
	}
	fmt.Println("Type of image:", t)

	//rectange of image
	rctSrc := img.Bounds()
	fmt.Println("Width:", rctSrc.Dx())
	fmt.Println("Height:", rctSrc.Dy())

	w := func() int {
		if params.Width != nil  {
			return *params.Width
		}
		return rctSrc.Dx()
	}()

	h := func() int {
		if params.Height != nil  {
			return *params.Height
		}
		return rctSrc.Dy()
	}()

	dst := new(bytes.Buffer)
	imgDst := imaging.Resize(img, w, h, imaging.Lanczos)

	//encode resized image
	/*
	dst := new(bytes.Buffer)
	switch t {
	case "jpeg":
		if err := jpeg.Encode(dst, imgDst, &jpeg.Options{Quality: 100}); err != nil {
			return
		}
	case "gif":
		if err := gif.Encode(dst, imgDst, nil); err != nil {
			return
		}
	case "png":
		if err := png.Encode(dst, imgDst); err != nil {
			return
		}
	default:
		fmt.Fprintln(os.Stderr, "format error")
	}

	 */
	err = imaging.Encode(dst , imgDst , imaging.JPEG)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return dst.Bytes() , nil

}


type Params struct {
	Width *int `mapstructure:"width"`
	Height *int  `mapstructure:"height"`
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


	err := mapstructure.Decode(m, &p)
	if err != nil {
		log.Error(err)
	}

	return p
}