package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/cavaliergopher/grab/v3/pkg/grabui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

//go:embed static
var static embed.FS

var (
	dir  *string
	lock sync.Mutex
	fp   string
)

type message map[string]string

func main() {
	dir = flag.String("d", "Download", "Save File Directory") //nolint
	Port := flag.String("p", "1018", "Run Port")
	flag.Parse()

	r := echo.New()
	r.Use(middleware.Recover())

	r.POST("download", Download)
	r.GET("state", State)

	// 处理静态文件
	resFs, _ := fs.Sub(static, "static")
	r.StaticFS("/web", resFs)
	err := r.Start(":" + *Port)
	if err != nil {
		log.Fatal("端口可能被占用")
	}
}

func State(c echo.Context) error {
	if lock.TryLock() {
		return c.JSON(http.StatusOK, message{
			"status":  "ok",
			"message": "已完成下载",
			"size":    getFileSize(fp),
		})
	}
	return c.JSON(http.StatusOK, message{
		"status":  "ok",
		"message": "正在下载中...",
		"size":    getFileSize(fp),
	})
}

func Download(c echo.Context) error {
	jsonStr := make(map[string]string) //注意该结构接受的内容
	if c.Bind(&jsonStr) != nil {
		return c.JSON(http.StatusBadRequest, message{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "参数不正确",
		})
	}

	u, err := url.Parse(jsonStr["link"])
	if err != nil {
		return c.JSON(http.StatusBadRequest, message{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "参数不正确",
		})
	}

	if !u.IsAbs() {
		return c.JSON(http.StatusBadRequest, message{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "下载地址不正确",
		})
	}

	fp = filepath.Join(*dir, filepath.Base(jsonStr["link"]))
	if e := download(jsonStr["link"], fp); e != nil {
		return c.JSON(http.StatusBadRequest, message{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "不支持的地址:" + e.Error(),
		})
	}
	return c.JSON(http.StatusOK, message{
		"status":  "ok",
		"message": "下载成功",
		"size":    getFileSize(fp),
	})
}

func download(url, dir string) error {
	lock.Lock()
	defer lock.Unlock()
	respch, err := grabui.GetBatch(context.Background(), 0, dir, url)
	if err != nil {
		return err
	}

	failed := 0
	for resp := range respch {
		if resp.Err() != nil {
			failed++
		}
		fmt.Println(resp.Size())
	}
	return nil
}

func getFileSize(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "0 MB"
	}
	return fmt.Sprintf("%d MB", fileInfo.Size()/1024/1024)
}
