package main

import (
	"github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
	_ "github.com/maczh/mgate/docs"
	"github.com/maczh/mgate/service"
	"github.com/maczh/mgin/errcode"
	"github.com/maczh/mgin/i18n"
	"github.com/maczh/mgin/middleware/cors"
	"github.com/maczh/mgin/middleware/postlog"
	"github.com/maczh/mgin/middleware/trace"
	"github.com/maczh/mgin/middleware/xlang"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"net/http"
	"strings"
)

/**
统一路由映射入口
*/
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	engine := gin.Default()

	//添加跟踪日志
	engine.Use(trace.TraceId())

	//设置接口日志
	engine.Use(postlog.RequestLogger())
	//添加跨域处理
	engine.Use(cors.Cors())
	//添加国际化支持
	engine.Use(xlang.RequestLanguage())

	//处理全局异常
	engine.Use(nice.Recovery(recoveryHandler))

	//设置404返回的内容
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, i18n.Error(errcode.URI_NOT_FOUND, errcode.UrlNotFound))
	})

	//从数据库加载网关API配置
	service.Gate.Init()

	//自动生成路由器
	engine.Any("/*action", func(c *gin.Context) {
		if service.Gate.GateConfig.Api.Swagger.Show {
			if c.Request.RequestURI == "/docs/doc.json" {
				c.JSON(http.StatusOK, service.Swagger.Get())
				return
			}
			if strings.HasPrefix(c.Request.RequestURI, "/docs/") {
				ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
				return
			}
		}
		if c.Request.RequestURI == "/favicon.ico" {
			c.String(http.StatusOK, "")
			return
		}
		if !service.Gate.CheckAuth(c) {
			c.JSON(http.StatusOK, i18n.Error(errcode.AUTHENTICATION_FAILURE, "签名错误，接口访问授权失败"))
			return
		}
		resp, err := service.Gate.ProxyTo(c)
		if err != nil {
			c.JSON(http.StatusOK, i18n.Error(errcode.SYSTEM_ERROR, err.Error()))
			return
		}
		c.String(http.StatusOK, resp)
		return
	})

	return engine
}

func recoveryHandler(c *gin.Context, err interface{}) {
	c.JSON(http.StatusOK, i18n.Error(errcode.SYSTEM_ERROR, errcode.SystemError))
}
