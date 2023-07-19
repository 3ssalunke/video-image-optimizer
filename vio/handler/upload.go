package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/3ssalunke/vio/vio/config"
	"github.com/3ssalunke/vio/vio/entity"
	"github.com/3ssalunke/vio/vio/utils"
	"github.com/labstack/echo/v4"
)

type UploadResult struct {
	Filename string `json:"filename"`
	Url      string `json:"url"`
}

func (v *VioHandler) UploadMultipart(c echo.Context) error {
	if _, err := v.checkSecretCode(c); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "secret code required")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	var uploaded []UploadResult

	for _, multipartHeader := range form.File["media"] {
		log.Printf("Processing file: %s", multipartHeader.Filename)

		bytes, err := multipartToBytes(multipartHeader)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		result, err := v.uploadBytes(multipartHeader.Filename, bytes)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		uploaded = append(uploaded, UploadResult{
			Filename: result.Filename,
			Url:      result.Url(),
		})
	}

	if len(uploaded) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No files to upload")
	}

	return renderUploadResults(uploaded, c)
}

func (v *VioHandler) UploadBodyBytes(c echo.Context) error {
	if _, err := v.checkSecretCode(c); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Secret code required")
	}

	var bytes []byte
	if c.Request().Body != nil {
		bytes, _ = ioutil.ReadAll(c.Request().Body)
	}

	result, err := v.uploadBytes("", bytes)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return renderUploadResults([]UploadResult{
		{Url: "/" + result.Filename},
	}, c)
}

func (v *VioHandler) uploadBytes(filename string, bytes []byte) (*entity.ProcessingFile, error) {
	file := &entity.ProcessingFile{
		Filename: filename,
		Mime:     utils.DetectMimeType(filename, bytes),
		Bytes:    bytes,
	}

	if file.IsImage() {
		log.Printf("Processing image: %s", file.Mime)
		err := utils.CalculateHashName(file)
		if err != nil {
			return file, err
		}

		if !config.App.Images.StoreOriginals {
			err = v.Processing.Image.AutoRotate(file)
			if err != nil {
				return file, err
			}

			err = v.Processing.Image.Resize(file, config.App.Images.OriginalLength)
			if err != nil {
				return file, err
			}

			if config.App.Images.AutoConvert != "false" {
				err = v.Processing.Image.Convert(file, config.App.Images.AutoConvert)
				if err != nil {
					return file, err
				}
			}
		}
	} else if file.IsVideo() {
		log.Printf("Processing video: %s", file.Mime)
		err := utils.CalculateHashName(file)
		if err != nil {
			return file, err
		}

		if !config.App.Videos.StoreOriginals {
			err = v.Processing.Video.Transcode(file, config.App.Videos.OriginalLength)
			if err != nil {
				return file, err
			}

			if config.App.Videos.AutoConvert != "false" {
				err = v.Processing.Video.Convert(file, config.App.Videos.AutoConvert)
				if err != nil {
					return file, err
				}
			}
		}
	} else {
		return nil, errors.New(fmt.Sprintf("unsupported file type: %s", file.Mime))
	}

	err := v.Storge.StoreFile(file, "orig")
	if err != nil {
		return file, err
	}

	return file, err
}

func multipartToBytes(multipartFile *multipart.FileHeader) ([]byte, error) {
	src, err := multipartFile.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return ioutil.ReadAll(src)
}

func renderUploadResults(results []UploadResult, c echo.Context) error {
	accept := c.Request().Header.Get("Accept")

	if strings.HasPrefix(accept, "application/json") {
		var urls []string
		for _, result := range results {
			urls = append(urls, result.Url)
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"uploaded": urls,
		})
	}

	var filenames []string
	for _, result := range results {
		filenames = append(filenames, result.Filename)
	}

	return c.Redirect(http.StatusFound, "/meta/"+strings.Join(filenames, ","))
}
