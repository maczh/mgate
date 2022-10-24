package service

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/levigross/grequests"
	"github.com/maczh/mgate/model"
	"github.com/maczh/mgin/client"
	"github.com/maczh/mgin/config"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/models"
	"github.com/maczh/mgin/utils"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"strings"
)

type gateConfig struct {
	GateConfig *model.GateConfig
}

var Gate = &gateConfig{
	GateConfig: new(model.GateConfig),
}

func (g *gateConfig) Init() {
	configFilePrefix := config.Config.GetConfigString("gate.config")
	configUrl := config.Config.GetConfigUrl(configFilePrefix)
	resp, err := grequests.Get(configUrl, &grequests.RequestOptions{})
	if err != nil {
		logs.Error("获取网关配置文件内容失败:{}", err.Error())
		return
	}
	err = yaml.Unmarshal(resp.Bytes(), g.GateConfig)
	if err != nil {
		logs.Error("网关配置解析失败:{}", err.Error())
		return
	}
	Swagger.Init(g)
	return
}

func (g *gateConfig) CheckAuth(c *gin.Context) bool {
	if !g.GateConfig.Api.Authorization.Need {
		return true
	}
	for _, gate := range g.GateConfig.Api.Gates {
		if strings.HasPrefix(c.Request.RequestURI, gate.Prefix) {
			subUri := c.Request.RequestURI[len(gate.Prefix):]
			if gate.Unauthorized != nil && len(gate.Unauthorized) > 0 {
				for _, path := range gate.Unauthorized {
					uri := path.Path
					matched := true
					if strings.Contains(path.Path, "**") {
						strs := strings.Split(path.Path, "**")
						for _, str := range strs {
							matched = matched && strings.Contains(subUri, str)
						}
					} else {
						matched = strings.HasPrefix(subUri, uri)
					}
					if matched && (path.Method == "" || path.Method == c.Request.Method) {
						return true
					}
				}
			}
		} else {
			continue
		}
	}
	authParams := make(map[string]string)
	for k, v := range g.GateConfig.Api.Authorization.Params {
		authParams[k] = c.GetHeader(v)
	}
	return g.callAuth(g.GateConfig.Api.Authorization.Service, g.GateConfig.Api.Authorization.Uri, authParams)
}

func (g *gateConfig) callAuth(service, uri string, params map[string]string) bool {
	logs.Debug("调用授权验证接口参数:{}", params)
	var resp string
	var err error
	if g.GateConfig.Api.Authorization.Method == "POST" {
		resp, err = client.Nacos.Call(service, uri, params)
	} else if g.GateConfig.Api.Authorization.Method == "GET" {
		resp, err = client.Get(service, uri, params)
	}
	if err != nil {
		logs.Error("授权验证接口调用失败:{}", err.Error())
		return false
	}
	var result models.Result
	utils.FromJSON(resp, &result)
	if result.Status != 1 {
		return false
	}
	data := make(map[string]interface{})
	utils.FromJSON(utils.ToJSON(result.Data), &data)
	if valid, ok := data[g.GateConfig.Api.Authorization.Response.Data]; ok {
		return valid.(bool)
	} else {
		return false
	}
}

func (g *gateConfig) ProxyTo(c *gin.Context) (string, error) {
	service, uri := g.getServiceUri(c)
	if service == "" || uri == "" {
		return "", errors.New("404 Not Found")
	}
	headers := getHeaders(c)
	params := utils.GinParamMap(c)
	body := make(map[string]interface{})
	c.ShouldBindJSON(&body)
	logs.Debug("headers:{},json:{}", headers, body)
	switch c.ContentType() {
	case "application/json":
		return client.Nacos.CallRestful(
			service,
			uri,
			c.Request.Method,
			getPathParams(c),
			params,
			headers,
			body,
		)
	case "application/x-www-form-urlencoded", "":
		switch c.Request.Method {
		case "POST":
			return client.CallWithHeader(
				service,
				uri,
				params,
				headers)
		case "GET":
			return client.GetWithHeader(
				service,
				uri,
				params,
				headers)
		}
	case "multipart/form-data":
		files := make([]grequests.FileUpload, 0)
		for k, v := range c.Request.MultipartForm.File {
			f, _, err := c.Request.FormFile(k)
			if err != nil {
				continue
			}
			data, err := ioutil.ReadAll(f)
			if err != nil {
				continue
			}
			file := grequests.FileUpload{
				FileName:     v[0].Filename,
				FileContents: io.NopCloser(bytes.NewReader(data)),
				FieldName:    k,
				FileMime:     "",
			}
			files = append(files, file)
		}
		return client.CallWithFilesHeader(service, uri, utils.GinParamMap(c), files, headers)
	}
	return "", errors.New("不支持" + c.ContentType() + "的Content-Type类型")
}

func (g *gateConfig) getServiceUri(c *gin.Context) (string, string) {
	for _, gate := range g.GateConfig.Api.Gates {
		if strings.HasPrefix(c.Request.RequestURI, gate.Prefix) {
			subUri := c.Request.RequestURI[len(gate.Prefix):]
			if subUri[:1] != "/" {
				subUri = "/" + subUri
			}
			for _, path := range gate.Allow {
				u := path.Path
				matched := true
				if strings.Contains(path.Path, "**") {
					strs := strings.Split(path.Path, "**")
					for _, str := range strs {
						matched = matched && strings.Contains(subUri, str)
					}
				} else {
					matched = strings.HasPrefix(subUri, u)
				}
				if matched && (path.Method == "" || path.Method == c.Request.Method) {
					for _, block := range gate.Block {
						u := block.Path
						m := true
						if strings.Contains(block.Path, "**") {
							strs := strings.Split(block.Path, "**")
							for _, str := range strs {
								m = m && strings.Contains(subUri, str)
							}
						} else {
							m = strings.HasPrefix(subUri, u)
						}
						if m && (block.Method == "" || block.Method == c.Request.Method) {
							return "", ""
						}
					}
					return gate.Service, subUri
				}
			}
			break
		} else {
			continue
		}
	}
	return "", ""
}

func getPathParams(c *gin.Context) map[string]string {
	p := make(map[string]string)
	for _, param := range c.Params {
		p[param.Key] = param.Value
	}
	return p
}

func getHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		headers[k] = v[0]
	}
	return headers
}
