package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maczh/logs"
	"github.com/maczh/mgconfig"
	"github.com/maczh/mgerr"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

//@title	通用微服务网关
//@version 	1.0.0(mgate)
//@description	通用微服务网关

//初始化命令行参数
func parseArgs() string {
	var configFile string
	flag.StringVar(&configFile, "f", os.Args[0]+".yml", "yml配置文件名")
	flag.Parse()
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if !strings.Contains(configFile, "/") {
		configFile = path + "/" + configFile
	}
	return configFile
}

func main() {
	//初始化配置，自动连接数据库和Nacos服务注册
	configFile := parseArgs()
	mgconfig.InitConfig(configFile)
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	//GIN的模式，生产环境可以设置成release
	gin.SetMode("debug")

	//国际化错误代码初始化
	mgerr.Init()

	engine := setupRouter()

	server := &http.Server{
		Addr:    ":" + mgconfig.GetConfigString("go.application.port"),
		Handler: engine,
	}
	serverSsl := &http.Server{
		Addr:    ":" + mgconfig.GetConfigString("go.application.port_ssl"),
		Handler: engine,
	}

	logs.Info("|-----------------------------------|")
	logs.Info("|      通用微服务网关MGate 1.0.0      |")
	logs.Info("|-----------------------------------|")
	logs.Info("|  Go Http Server Start Successful  |")
	logs.Info("|    Port:" + mgconfig.GetConfigString("go.application.port") + "     Pid:" + fmt.Sprintf("%d", os.Getpid()) + "        |")
	logs.Info("|-----------------------------------|")
	logs.Info("")

	if mgconfig.GetConfigString("go.application.port") != "" {
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logs.Error("HTTP server listen: " + err.Error())
			}
		}()
	}

	if mgconfig.GetConfigString("go.application.cert") != "" {
		go func() {
			var err error
			err = serverSsl.ListenAndServeTLS(path+"/"+mgconfig.GetConfigString("go.application.cert"), path+"/"+mgconfig.GetConfigString("go.application.key"))
			if err != nil && err != http.ErrServerClosed {
				logs.Error("HTTPS server listen: {}", err.Error())
			}
		}()
	}

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signalChan
	logs.Error("Get Signal:" + sig.String())
	logs.Error("Shutdown Server ...")
	mgconfig.SafeExit()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logs.Error("Server Shutdown:" + err.Error())
	}
	logs.Error("Server exiting")

}
