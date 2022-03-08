package main

import (
	"github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
	"github.com/maczh/gintool"
	"github.com/maczh/gintool/mgresult"
	"github.com/maczh/mgtrace"
	"github.com/maczh/utils"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"mgate/docs"
	"mgate/service"
	"net/http"
)

/**
统一路由映射入口
*/
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	engine := gin.Default()

	//添加跟踪日志
	engine.Use(mgtrace.TraceId())

	//设置接口日志
	engine.Use(gintool.SetRequestLogger())
	//添加跨域处理
	engine.Use(gintool.Cors())

	//处理全局异常
	engine.Use(nice.Recovery(recoveryHandler))

	//设置404返回的内容
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, mgresult.Error(-1, "请求的方法不存在"))
	})

	//从数据库加载网关API配置
	service.LoadDataFromMongoDB()
	//自动生成路由器
	service.GenerateRoutes(engine)
	//swag初始化
	docs.Init()

	//获取Swagger Json
	//engine.GET("/docs/doc.json", func(c *gin.Context) {
	//	c.JSON(http.StatusOK, service.GetApiDocsJson())
	//})

	//添加swagger支持
	engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//添加管理接口
	engine.POST("/admin/add/swagger", func(c *gin.Context) {
		params := utils.GinParamMap(c)
		c.JSON(http.StatusOK, service.AddApiWithSwagger(params["apiPath"], params["service"], params["uri"], params["withHeader"], engine))
	})

	engine.POST("/admin/add/api", func(c *gin.Context) {
		params := utils.GinParamMap(c)
		c.JSON(http.StatusOK, service.AddApi(params["apiPath"],
			params["service"],
			params["uri"],
			params["withHeader"],
			params["method"],
			params["description"],
			params["summary"],
			params["consume"],
			params["produce"],
			params["tag"],
			params["parameters"],
			engine))
	})

	return engine
}

func recoveryHandler(c *gin.Context, err interface{}) {
	c.JSON(http.StatusOK, mgresult.Error(-1, "系统异常，请联系客服"))
}
