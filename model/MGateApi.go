package model

import "gopkg.in/mgo.v2/bson"

type ServiceApi struct {
	Api     string `json:"api" bson:"api"`         //网关映射的API地址
	Service string `json:"service" bson:"service"` //微服务名
	Uri     string `json:"uri" bson:"uri"`         //微服务URI地址
}

type GatewayApi struct {
	ServiceApi
	Swagger ApiDocument `json:"swagger" bson:"swagger"` //接口Swagger文档
}

type GatewayApiDocument struct {
	Id bson.ObjectId `json:"id" bson:"_id"`
	GatewayApi
}
