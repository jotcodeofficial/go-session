package gosession

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

/*
	S3 FILE SYSTEM

	The s3 file system is used to interact with the clusters s3 integration service.
	It will query the service and provide the correct authentication details along with the s3File
	request object. If successful it should store the file and get the valid URL back to store
	in the database.

	It should also expose an api to backup/restore the buckets for applications this api has permissions for.

*/

// SendS3File struct
type SendS3File struct {
	FileType        string `json:"fileType,omitempty" bson:"fileType,omitempty" validate:"required"`
	FolderStructure string `json:"folderStructure,omitempty" bson:"folderStructure,omitempty" validate:"required"`
	ProjectName     string `json:"projectName,omitempty" bson:"projectName,omitempty" validate:"required"`
	FileName        string `json:"fileName,omitempty" bson:"fileName,omitempty" validate:"required"`
	//File            *multipart.File `json:"file,omitempty" bson:"file,omitempty" validate:"required"`
	File []byte `json:"file,omitempty" bson:"file,omitempty" validate:"required"`
}

// ReturnedS3File struct
type ReturnedS3File struct {
	FileName string `json:"fileName,omitempty" bson:"fileName,omitempty" validate:"required"`
	URL      string `json:"url,omitempty" bson:"url,omitempty" validate:"required, url"`
}

// Configuration Section --------------------------------------------

func configureS3Routes() {

	e.POST("/user/upload-file", uploadFileToS3)

}

// endpoint for sending a file to s3-service and returning a url
func uploadFileToS3(c echo.Context) error {

	// this is the name of the file
	name := c.FormValue("name")
	// this is the email of the submitting user
	email := c.FormValue("email")

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	// dst, err := os.Create(file.Filename)
	// if err != nil {
	// 	return err
	// }
	// defer dst.Close()

	// TODO change this to send it via POST to s3 service
	// Copy
	// if _, err = io.Copy(dst, src); err != nil {
	// 	return err
	// }
	byteContainer, err := ioutil.ReadAll(src)
	if err != nil {
		fmt.Println(err)
	}

	s3File := SendS3File{}
	//s3File.File = bytes.NewBuffer(&src)
	s3File.File = byteContainer
	s3File.FileName = "test"
	s3File.FileType = "PNG"
	s3File.FolderStructure = "/"
	s3File.ProjectName = "project1"

	byteInfo, err := json.Marshal(s3File)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusNotFound, err.Error())
	}
	// create request to the s3 service
	// TODO update endpoint in production
	resp, err := http.NewRequest("POST", "http://127.0.0.1:8082/send-s3-file", bytes.NewBuffer(byteInfo))
	if err != nil {
		fmt.Println("error creating the post request")
		fmt.Println(err)
		return c.String(http.StatusNotFound, err.Error())
	}

	client := &http.Client{}
	newResp, err := client.Do(resp)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusNotFound, err.Error())
	}

	defer newResp.Body.Close()
	body, err := ioutil.ReadAll(newResp.Body)
	if err != nil {
		fmt.Println("return from the post request to s3 service")
		fmt.Println(string(body))
		fmt.Println(err)
		return c.String(http.StatusNotFound, err.Error())
	}

	// "file":  file.Filename,
	// "name":  name,
	// "email": email,
	fmt.Println(file.Filename)
	fmt.Println(name)
	fmt.Println(email)

	// TODO when file is uploaded save here as random name so not to have others

	// TODO save the url to the database for that in the part of the api that calls this function (internal function not endpoint)

	// TODO delete the test.png after having sen it to s3-service

	return c.JSON(http.StatusOK, ReturnedS3File{
		FileName: "TestName.png",
		URL:      "https://domain.com",
	})
}
