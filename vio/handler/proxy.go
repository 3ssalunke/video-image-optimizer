package handler

import (
	"errors"
	"net/http"
	"path"
	"strconv"

	"github.com/3ssalunke/vio/vio/config"
	"github.com/3ssalunke/vio/vio/entity"
	"github.com/labstack/echo/v4"
)

const minLength = 200

func (v *VioHandler) GetOriginalFile(c echo.Context) error {
	file, err := v.Storge.GetFile("orig", c.Param("name"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "File not found")
	}

	return v.Storge.Proxy(c, file.Path)
}

func (v *VioHandler) GetResizedFile(c echo.Context) error {
	lengthString := c.Param("length")
	if lengthString == "full" {
		return v.GetOriginalFile(c)
	}

	length, err := strconv.Atoi(lengthString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad 'length' parameter. Need an integer!")
	}

	if length > config.App.Images.OriginalLength {
		return v.GetOriginalFile(c)
	}

	if length < minLength {
		length = minLength
	}

	filename := c.Param("name")
	file, err := v.resizeFile(filename, length)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return v.Storge.Proxy(c, file.Path)
}

func (v *VioHandler) resizeFile(filename string, length int) (*entity.ProcessingFile, error) {
	resizePath := path.Join("resize", strconv.Itoa(length))
	file, err := v.Storge.GetFile(resizePath, filename)
	if err == nil {
		return file, nil
	}

	if file.IsImage() {
		if config.App.Images.LiveResize {
			err := v.Storge.ReadFileBytes(file, "orig")
			if err != nil {
				return file, err
			}

			err = v.Processing.Image.Resize(file, length)
			if err != nil {
				return file, err
			}

			err = v.Storge.StoreFile(file, resizePath)
			if err != nil {
				return file, err
			}
		}

		return file, nil
	}
	if file.IsVideo() {
		if config.App.Videos.LiveResize {
			err := v.Storge.ReadFileBytes(file, "orig")
			if err != nil {
				return file, err
			}

			err = v.Processing.Video.Transcode(file, length)
			if err != nil {
				return file, err
			}

			err = v.Storge.StoreFile(file, resizePath)
			if err != nil {
				return file, err
			}
		}

		return file, nil
	}

	return nil, errors.New("file does not exist")
}
