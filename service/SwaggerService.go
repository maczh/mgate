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

var swaggerDocuments model.SwaggerDocument

func AddApiWithSwagger(apiPath, service, uri string, engine *gin.Engine) mgresult.Result {
	resp, err := mgcall.Get(service, "/docs/doc.json", nil)
	if err != nil {
		logs.Error("{}的swagger文档无法访问")
		return mgresult.Error(-1, service+"的swagger文档无法访问")
	}
	swaggerDocs := model.SwaggerDocument{}
	utils.FromJSON(resp, &swaggerDocs)
	if doc, ok := swaggerDocs.Paths[uri]; ok {
		gatewayApi := model.GatewayApi{
			ServiceApi: model.ServiceApi{
				Api:     apiPath,
				Service: service,
				Uri:     uri,
			},
			Swagger: doc,
		}
		mgateApiInfo[apiPath] = gatewayApi
		//添加swagger路径
		swaggerDocuments.Paths[apiPath] = gatewayApi.Swagger
		//保存入库
		mgoDoc := model.GatewayApiDocument{}
		err = mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).Find(bson.M{"api": apiPath}).One(&mgoDoc)
		if err != nil || mgoDoc.Api == "" {
			mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).Insert(gatewayApi)
		} else {
			mgoDoc.GatewayApi = gatewayApi
			mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).UpdateId(mgoDoc.Id, mgoDoc)
		}
		//动态添加路由
		if _, f := doc["post"]; f {
			engine.POST(apiPath, func(c *gin.Context) {
				result := Route(c)
				c.JSON(http.StatusOK, result)
			})
		} else if _, f = doc["get"]; f {
			engine.GET(apiPath, func(c *gin.Context) {
				result := Route(c)
				c.JSON(http.StatusOK, result)
			})
		}
		return mgresult.Success(nil)
	} else {
		return mgresult.Error(-1, "添加新网关路由失败：目标服务接口不存在")
	}
}

func AddApi(apiPath, service, uri, method, description, summary, consume, produce, tag, parameters string, engine *gin.Engine) mgresult.Result {
	if method == "" {
		method = "post"
	}
	if consume == "" {
		consume = "application/x-www-form-urlencoded"
	}
	if produce == "" {
		produce = "application/json"
	}
	var params []model.ApiParameter
	{
	}
	utils.FromJSON(parameters, &params)
	doc := model.ApiInfo{
		Description: description,
		Consumes:    []string{consume},
		Produces:    []string{produce},
		Tags:        []string{tag},
		Summary:     summary,
		Parameters:  params,
	}
	doc.Responses.Field1.Description = "ok"
	doc.Responses.Field1.Schema.Type = "string"
	gatewayApi := model.GatewayApi{
		ServiceApi: model.ServiceApi{
			Api:     apiPath,
			Service: service,
			Uri:     uri,
		},
		Swagger: model.ApiDocument{method: doc},
	}
	mgateApiInfo[apiPath] = gatewayApi
	//添加swagger路径
	swaggerDocuments.Paths[apiPath] = gatewayApi.Swagger
	//保存入库
	mgoDoc := model.GatewayApiDocument{}
	err := mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).Find(bson.M{"api": apiPath}).One(&mgoDoc)
	if err != nil || mgoDoc.Api == "" {
		mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).Insert(gatewayApi)
	} else {
		mgoDoc.GatewayApi = gatewayApi
		mgconfig.Mgo.C(mgconfig.GetConfigString("api.collection")).UpdateId(mgoDoc.Id, mgoDoc)
	}
	//动态添加路由
	if method == "post" {
		engine.POST(apiPath, func(c *gin.Context) {
			result := Route(c)
			c.JSON(http.StatusOK, result)
		})
	} else if method == "get" {
		engine.GET(apiPath, func(c *gin.Context) {
			result := Route(c)
			c.JSON(http.StatusOK, result)
		})
	}
	return mgresult.Success(nil)
}

func GetApiDocsJson() model.SwaggerDocument {
	return swaggerDocuments
}

func addAdminApiSwaggerDocs() {
	//添加 AddApiWithSwagger 接口
	params := []model.ApiParameter{
		model.ApiParameter{
			Type:        "string",
			Description: "网关接口路径",
			Name:        "apiPath",
			In:          "formData",
			Required:    true,
		},
		model.ApiParameter{
			Type:        "string",
			Description: "微服务名称",
			Name:        "service",
			In:          "formData",
			Required:    true,
		},
		model.ApiParameter{
			Type:        "string",
			Description: "在微服务端接口路径",
			Name:        "uri",
			In:          "formData",
			Required:    true,
		},
	}
	doc := model.ApiInfo{
		Description: "从微服务swagger添加一个网关映射",
		Consumes:    []string{"application/x-www-form-urlencoded"},
		Produces:    []string{"application/json"},
		Tags:        []string{"网关管理"},
		Summary:     "从带swagger文档的微服务网关添加一个接口映射",
		Parameters:  params,
	}
	doc.Responses.Field1.Description = "ok"
	doc.Responses.Field1.Schema.Type = "string"
	swaggerDocuments.Paths["/admin/add/swagger"] = model.ApiDocument{"post": doc}
	//添加AddApi接口
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "http方法，post或get",
		Name:        "method",
		In:          "formData",
		Required:    false,
	})
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "接口详情描述",
		Name:        "description",
		In:          "formData",
		Required:    true,
	})
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "接口描述简述",
		Name:        "summary",
		In:          "formData",
		Required:    true,
	})
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "接口content-type,默认为application/x-www-form-urlencoded",
		Name:        "consume",
		In:          "formData",
		Required:    false,
	})
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "接口返回content-Type,默认为application/json",
		Name:        "produce",
		In:          "formData",
		Required:    false,
	})
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "分组标签",
		Name:        "tag",
		In:          "formData",
		Required:    true,
	})
	params = append(params, model.ApiParameter{
		Type:        "string",
		Description: "接口参数JSON,包含name,type,in,required,description等5个参数",
		Name:        "parameters",
		In:          "formData",
		Required:    true,
	})
	doc1 := model.ApiInfo{
		Description: "直接添加一个网关映射",
		Consumes:    []string{"application/x-www-form-urlencoded"},
		Produces:    []string{"application/json"},
		Tags:        []string{"网关管理"},
		Summary:     "微服务网关直接添加一个接口映射",
		Parameters:  params,
	}
	doc1.Responses.Field1.Description = "ok"
	doc1.Responses.Field1.Schema.Type = "string"
	swaggerDocuments.Paths["/admin/add/api"] = model.ApiDocument{"post": doc1}
}
