package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/so-brian/file-transfer-service/internal/pkg/azure/storage"
	"github.com/so-brian/file-transfer-service/internal/pkg/utility"
)

type UploadRequest struct {
	Session string `uri:"key" binding:"required"`
}

type CacheRequest struct {
	Key    string    `json:"key"`
	Value  string    `json:"value"`
	Expire time.Time `json:"expire"`
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// create a file transfer session
	r.POST("/session", func(c *gin.Context) {
		// create a session
		key := utility.RandStr(6)
		expire := time.Now().UTC().Add(24 * time.Hour)
		body := CacheRequest{
			Key:    key,
			Value:  "",
			Expire: expire,
		}

		// Create a buffer to store the JSON encoded data
		buffer := new(bytes.Buffer)

		// Create a new JSON encoder using the buffer
		encoder := json.NewEncoder(buffer)

		// Encode the request body to JSON
		if err := encoder.Encode(body); err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}

		res, err := http.DefaultClient.Post("https://apim.sobrian.net/cache-service/string", "application/json", buffer)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": err})
			return
		}

		if res.StatusCode != http.StatusCreated {
			c.JSON(res.StatusCode, gin.H{"msg": res.Status})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"key": key, "expire": expire})
	})

	// delete a file transfer session
	r.DELETE("/session/:key", func(c *gin.Context) {
		var req UploadRequest
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		uri := fmt.Sprintf("https://sobrian.blob.core.windows.net/%s", req.Session)

		// test if the session exist
		res, err := http.DefaultClient.Get(uri)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		if res.StatusCode != http.StatusOK {
			c.JSON(res.StatusCode, gin.H{"msg": res.Status})
			return
		}

		// delete the session
		deleteReq := http.Request{
			Method: http.MethodDelete,
			URL:    &url.URL{Path: uri},
		}

		res, err = http.DefaultClient.Do(&deleteReq)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		if res.StatusCode != http.StatusOK {
			c.JSON(res.StatusCode, gin.H{"msg": res.Status})
			return
		}

		// delete files
		client := storage.NewAzureStorageClient()
		err = client.DeleteGroup(req.Session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "deleted"})
	})

	// upload a file to a session
	r.POST("/session/:key", func(c *gin.Context) {
		var req UploadRequest
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		// test if the session exist
		res, err := http.DefaultClient.Get(fmt.Sprintf("https://apim.sobrian.net/cache-service/string/%s", req.Session))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		if res.StatusCode != http.StatusOK {
			c.JSON(res.StatusCode, gin.H{"msg": res.Status})
			return
		}

		// single file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		log.Println(file.Filename)
		content, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		// Upload the file to specific dst.
		client := storage.NewAzureStorageClient()
		client.Upload(storage.File{
			Group:   req.Session,
			Name:    file.Filename,
			Content: content,
		})

		c.JSON(http.StatusCreated, gin.H{"msg": fmt.Sprintf("'%s' uploaded!", file.Filename)})
	})

	// get the list of files in a session
	r.GET("/session/:key", func(c *gin.Context) {
		var req UploadRequest
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		// test if the session exist
		res, err := http.DefaultClient.Get(fmt.Sprintf("https://apim.sobrian.net/cache-service/string/%s", req.Session))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		if res.StatusCode != http.StatusOK {
			c.JSON(res.StatusCode, gin.H{"msg": res.Status})
			return
		}

		// get the list of files
		client := storage.NewAzureStorageClient()
		files, err := client.GetList(req.Session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{"files": files})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
