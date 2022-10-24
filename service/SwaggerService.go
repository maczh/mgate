package service

import (
	"github.com/maczh/mgate/model"
	"github.com/maczh/mgin/client"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/utils"
	"strings"
)

type swaggerService struct {
	Doc *model.SwaggerDocument
}

var Swagger = &swaggerService{
	Doc: new(model.SwaggerDocument),
}

func (s *swaggerService) Init(g *gateConfig) {
	if !g.GateConfig.Api.Swagger.Show {
		return
	}
	s.Doc.Title = g.GateConfig.Api.Swagger.Title
	s.Doc.Description = g.GateConfig.Api.Swagger.Description
	s.Doc.Version = g.GateConfig.Api.Swagger.Version
	s.Doc.Info.Title = s.Doc.Title
	s.Doc.Info.Description = s.Doc.Description
	s.Doc.Info.Version = s.Doc.Version
	s.Doc.Swagger = "2.0"
	for service, _ := range g.GateConfig.Api.Gates {
		s.addServiceSwagger(service, g)
	}
	logs.Info("swagger文档自动导入完成")
}

func (s *swaggerService) Get() model.SwaggerDocument {
	return *s.Doc
}

func (s *swaggerService) addServiceSwagger(service string, g *gateConfig) {
	resp, err := client.Get(service, "/docs/doc.json", nil)
	if err != nil {
		logs.Error("{}的swagger文档无法访问", service)
		return
	}
	swaggerDocs := model.SwaggerDocument{}
	utils.FromJSON(resp, &swaggerDocs)
	//导入模型部分
	if s.Doc.Definitions == nil && swaggerDocs.Definitions != nil {
		s.Doc.Definitions = make(map[string]model.Model)
	}
	if swaggerDocs.Definitions != nil {
		for k, v := range swaggerDocs.Definitions {
			s.Doc.Definitions[k] = v
		}
	}
	//导入路径数据
	if s.Doc.Paths == nil {
		s.Doc.Paths = make(map[string]model.ApiDocument)
	}
	cfg := g.GateConfig.Api.Gates[service]
	tagsMap := g.GateConfig.Api.Swagger.Tags[service]
	for path, api := range swaggerDocs.Paths {
		method := getSwaggerApiMethod(api)
		if checkPath(path, method, cfg.Allow, cfg.Block) {
			doc := api[method]
			//添加认证参数
			if g.GateConfig.Api.Authorization.Need && !checkPath(path, method, cfg.Unauthorized, nil) {
				if doc.Parameters == nil {
					doc.Parameters = make([]model.ApiParameter, 0)
				}
				for param, _ := range g.GateConfig.Api.Authorization.Params {
					has := false
					for _, p := range doc.Parameters {
						if param == p.Name {
							has = true
							break
						}
					}
					if !has {
						doc.Parameters = append(doc.Parameters, model.ApiParameter{
							Type:        "string",
							Description: param + "- 认证参数",
							Name:        param,
							In:          "header",
							Required:    true,
						})
					}
				}
			}
			//修改tags标签
			for i, tag := range doc.Tags {
				if tagsMap != nil && tagsMap[tag] != "" {
					doc.Tags[i] = tagsMap[tag]
				}
			}
			api[method] = doc
			s.Doc.Paths[strings.TrimSuffix(cfg.Prefix, "/")+path] = api
			logs.Info("导入{}的Swagger文档路径{}成功", service, path)
		}
	}
}

func checkPath(path, method string, allows, blocks []model.PathConfig) bool {
	method = strings.ToUpper(method)
	allowed := false
	for _, allow := range allows {
		match := true
		p := allow.Path
		if strings.Contains(p, "**") {
			strs := strings.Split(p, "**")
			for _, str := range strs {
				match = match && strings.Contains(path, str)
			}
		} else {
			match = strings.HasPrefix(path, p)
		}
		if match && (allow.Method == "" || strings.ToUpper(allow.Method) == method) {
			allowed = true
			break
		}
	}
	if !allowed {
		return false
	}
	if blocks == nil || len(blocks) == 0 {
		return true
	}
	for _, block := range blocks {
		match := true
		p := block.Path
		if strings.Contains(p, "**") {
			strs := strings.Split(p, "**")
			for _, str := range strs {
				match = match && strings.Contains(path, str)
			}
		} else {
			match = strings.HasPrefix(path, p)
		}
		if match && (block.Method == "" || strings.ToUpper(block.Method) == method) {
			return false
		}
	}
	return true
}

func getSwaggerApiMethod(api model.ApiDocument) string {
	for method, _ := range api {
		return method
	}
	return ""
}
