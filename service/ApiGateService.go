package service

import (
	"github.com/gin-gonic/gin"
	"github.com/maczh/gintool/mgresult"
	"github.com/maczh/logs"
	"github.com/maczh/mgcall"
	"github.com/maczh/mgconfig"
	"github.com/maczh/utils"
	"gopkg.in/mgo.v2/bson"
	"mgate/model"
	"net/http"
)

var mgateApiInfo = make(map[string]model.GatewayApi)

func LoadDataFromMongoDB() {
	var apiData []model.GatewayApi
	query := bson.M{}
	err := mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).Find(query).All(&apiData)
	if err != nil {
		logs.Error("mongodb错误:{}", err.Error())
		return
	}
	swaggerDocuments.Swagger = "2.0"
	swaggerDocuments.Info.Description = mgconfig.GetConfigString("swagger.info.description")
	swaggerDocuments.Info.Title = mgconfig.GetConfigString("swagger.info.title")
	swaggerDocuments.Info.Version = mgconfig.GetConfigString("swagger.info.version")
	swaggerDocuments.Paths = make(map[string]model.ApiDocument)
	for _, api := range apiData {
		mgateApiInfo[api.Api] = api
		swaggerDocuments.Paths[api.Api] = api.Swagger
	}
	//添加管理接口
	addAdminApiSwaggerDocs()
	logs.Debug("API数据载入完成")
}

func Route(c *gin.Context) mgresult.Result {
	apiPath := c.Request.RequestURI
	params := utils.GinParamMap(c)
	resp := ""
	var err error
	if api, ok := mgateApiInfo[apiPath]; ok {
		if _, f := api.Swagger["post"]; f {
			resp, err = mgcall.Call(api.Service, api.Uri, params)
		} else if _, f = api.Swagger["get"]; f {
			resp, err = mgcall.Get(api.Service, api.Uri, params)
		}
	}
	if err != nil {
		return *mgresult.Error(-1, apiPath+"路由失败:"+err.Error())
	}
	var result mgresult.Result
	utils.FromJSON(resp, &result)
	return result
}

func GenerateRoutes(engine *gin.Engine) {
	for uri, api := range mgateApiInfo {
		if _, f := api.Swagger["post"]; f {
			engine.POST(uri, func(c *gin.Context) {
				result := Route(c)
				c.JSON(http.StatusOK, result)
			})
		} else if _, f = api.Swagger["get"]; f {
			engine.GET(uri, func(c *gin.Context) {
				result := Route(c)
				c.JSON(http.StatusOK, result)
			})
		}
	}
}
