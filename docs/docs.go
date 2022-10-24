// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"github.com/maczh/mgate/service"
	"github.com/maczh/mgin/utils"
	"github.com/swaggo/swag"
)

var apiJson = service.Swagger.Get()

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     apiJson.Version,
	Host:        apiJson.Host,
	BasePath:    apiJson.BasePath,
	Schemes:     []string{},
	Title:       apiJson.Title,
	Description: apiJson.Description,
}

type s struct{}

func (s *s) ReadDoc() string {
	return utils.ToJSON(service.Swagger.Get())
}

func Init() {
	swag.Register("swagger", &s{})
}
