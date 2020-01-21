package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zboyco/huaban/controller"
	"github.com/zboyco/huaban/model"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

func main() {

	fmt.Println("***********************告知**********************\n")
	for i := 0; i < 3; i++ {
		fmt.Println("          程序运行过程中，请勿关闭该窗口          \n")
	}
	fmt.Println("***********************告知**********************\n")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startWeb()
	}()

	go func() {
		for i := 5; i > 0; i-- {
			fmt.Println(fmt.Sprintf("程序将在%v秒后打开界面...", i))
			time.Sleep(time.Second)
		}

		fmt.Println("\n正在打开网页，如果没有自动打开，请手动打开此网页： http://localhost:9010\n")

		err := open("http://localhost:9010")
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}

func startWeb() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.LoadHTMLGlob("ui.html")

	msg := &model.Message{}
	msg.Reset()

	var ctx context.Context
	var cancel context.CancelFunc

	r.POST("/api/start", func(c *gin.Context) {
		body := &model.Body{}
		err := c.BindJSON(body)
		if err != nil {
			c.JSON(http.StatusBadRequest, nil)
			return
		}

		userAgent := c.Request.Header.Get("User-Agent")
		msg.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		go controller.StartDownload(ctx, body.Url, userAgent, msg)
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
		c.HTML(http.StatusOK, "ui.html", nil)
	})

	err := r.Run(":9010")
	if err != nil {
		fmt.Print(err)
	}
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
