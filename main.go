package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	_ "download/statik"
	"github.com/cavaliercoder/grab"
	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
)

var (
	Dir  string
	Resp *grab.Response
)

func main() {
	Dir = *flag.String("d", "/mnt/share/Download", "Save File Directory") //nolint
	Port := flag.String("p", "1018", "Run Port")
	flag.Parse()

	statikFS, _ := fs.New()
	r := gin.Default()
	r.POST("download", Download)
	r.GET("state", State)

	r.StaticFS("/web", statikFS)
	r.Run(":" + *Port)
	log.Fatal("端口被占用")
}

func State(c *gin.Context) {
	if Resp == nil {
		c.JSON(http.StatusOK, gin.H{
			"size":     "",
			"status":   "OK",
			"message":  "无下载任务",
			"filename": "",
		})
		return
	}
	if Resp.IsComplete() {
		if Resp.Err() != nil {
			c.JSON(Resp.HTTPResponse.StatusCode, gin.H{
				"status":  Resp.HTTPResponse.Status,
				"message": "下载失败",
				"error":   Resp.Err().Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"size":     Unwrap(Resp.BytesComplete()),
			"status":   "OK",
			"message":  "下载完成",
			"filename": filepath.Base(Resp.Filename),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"size":     Unwrap(Resp.BytesComplete()),
		"status":   "OK",
		"message":  "下载中...",
		"filename": filepath.Base(Resp.Filename),
	})
}

func Download(c *gin.Context) {
	jsonStr := make(map[string]string) //注意该结构接受的内容
	if c.BindJSON(&jsonStr) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "参数不正确",
		})
		return
	}

	u, err := url.Parse(jsonStr["link"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "参数不正确",
		})
		return
	}

	if !u.IsAbs() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "下载地址的URL格式不正确",
		})
		return
	}

	if Resp != nil {
		if !Resp.IsComplete() {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusText(http.StatusBadRequest),
				"message": "仍有下载任务未完成, 不支持多任务下载",
			})
			return
		}
	}

	client := grab.NewClient()
	req, err := grab.NewRequest(Dir, jsonStr["link"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "不支持的地址",
			"error":   err.Error(),
		})
		return
	}
	go func() {
		Resp = client.Do(req)
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "开始下载",
	})
}

func Unwrap(num int64) string {
	f := float64(num) / 1024 / 1024
	return fmt.Sprintf("%.2fM", f)
}
