package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/browser"
	"github.com/zboyco/huaban/controller"
	"github.com/zboyco/huaban/model"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

func main() {
	for i:= 0 ; i < 3;i++ {
		fmt.Println("告知：程序运行过程中，请勿关闭该窗口...\n")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startWeb()
	}()

	fmt.Println("正在打开网页，如果没有自动打开，请手动打开此网页： http://localhost:9010\n")

	err := browser.OpenURL("http://localhost:9010")
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}

func startWeb() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Static("/static", "public")
	r.LoadHTMLGlob("templates/*")

	msg := &model.Message{}
	msg.Reset()

	var ctx context.Context
	var cancel context.CancelFunc

	r.POST("/api/start", func(c *gin.Context) {
		data, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, nil)
			return
		}
		urlString := string(data)
		urlString, err = url.PathUnescape(urlString[1:])
		if err != nil {
			c.JSON(http.StatusBadRequest, nil)
			return
		}
		userAgent := c.Request.Header.Get("User-Agent")
		msg.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		go controller.StartDownload(ctx, urlString, userAgent, msg)
		c.JSON(http.StatusOK, nil)
	})
	r.POST("/api/pause", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})
	r.POST("/api/stop", func(c *gin.Context) {
		cancel()
		c.JSON(http.StatusOK, nil)
	})
	r.GET("/api/message", func(c *gin.Context) {
		result := msg.Pick()

		c.JSON(http.StatusOK, gin.H{
			"msgs": result,
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.Run(":9010")
}
