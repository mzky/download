package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/melbahja/got"
	"io/fs"
	"log"
	"net/http"
	"net/url"
)

//go:embed static
var static embed.FS

var (
	Dir  string
	Resp *grab.Response
)

type message map[string]string

func main() {
	Dir = *flag.String("d", "Download", "Save File Directory") //nolint
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
	return c.JSON(http.StatusOK, message{
		"status":  "ok",
		"message": "下载成功111",
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

	g := got.New()
	if e := g.Download(jsonStr["link"], Dir); e != nil {
		return c.JSON(http.StatusBadRequest, message{
			"status":  http.StatusText(http.StatusBadRequest),
			"message": "不支持的地址:" + e.Error(),
		})
	}
	return c.JSON(http.StatusOK, message{
		"status":  "ok",
		"message": "下载成功",
	})
}

func unwrap(num int64) string {
	f := float64(num) / 1024 / 1024
	return fmt.Sprintf("%.2fM", f)
}
