package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zboyco/huaban/controller"
	"github.com/zboyco/huaban/model"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

//const uiName string = "ui.html"

func main() {

	fmt.Println("***********************告知***********************\n")
	for i := 0; i < 3; i++ {
		fmt.Println("          程序运行过程中，请勿关闭该窗口          \n")
	}
	fmt.Println("***********************告知***********************\n")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startWeb()
	}()

	go func() {
		for i := 5; i > 0; i-- {
			fmt.Println(fmt.Sprintf("将在%v秒后打开界面...", i))
			time.Sleep(time.Second)
		}

		fmt.Println("\n正在打开网页，如果没有自动打开，请手动打开此网页： https://function.work/huaban/\n")

		err := open("https://function.work/huaban/")
		if err != nil {
			fmt.Println(err)
		}
	}()

	time.Sleep(6 * time.Second)

	fmt.Println("***********************告知***********************\n")
	for i := 0; i < 3; i++ {
		fmt.Println("          程序运行过程中，请勿关闭该窗口          \n")
	}
	fmt.Println("***********************告知***********************\n")

	wg.Wait()
}

func startWeb() {
	gin.SetMode(gin.ReleaseMode)

	gin.DefaultWriter = model.NilWriter{}

	r := gin.Default()
	//r.LoadHTMLGlob(uiName)

	r.Use(Cors())

	msg := &model.Message{}
	msg.Reset()

	var ctx context.Context
	var cancel context.CancelFunc

	r.POST("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

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
	//r.GET("/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, uiName, nil)
	//})

	err := r.Run(":9010")
	if err != nil {
		fmt.Print(err)
	}
}

// 跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "https://function.work")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			//c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
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
