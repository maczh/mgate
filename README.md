# MGate通用微服务网关说明

mgate是一款基于Go语言开发的，以Nacos作为注册中心与配置中心的，具备动态添加微服务接口转发配置的通用型微服务网关。

## 基本功能

- 采用yaml格式配置，配置文件存放在nacos配置中心
- 支持x-form与restful格式接口，支持multi-part文件上传中转
- 支持以header中参数通过微服务内部授权认证接口进行动态认证
- 支持同时外放多个微服务的接口,不同服务的接口地址前可分别加不同的路径前缀
- 支持按uri前缀与模糊匹配两种模式筛选允许外放的接口、阻止外族的接口与免认证接口
- 支持通过配置动态生成swagger与关闭swagger文档，支持swagger的分组tag动态修改

## 应用本地配置文件

```yaml
go:
  application:
    name: mgate-项目名
    port: 8087
    debug: false
#    port_ssl: 8089
#    cert: certs/gate.pem
#    key: certs/gate.key
  config:
    server: http://xxx.xxx.xxx.xxx:8848/
    server_type: nacos
    env: test
    type: .yml
    used: nacos,mongodb
    prefix:
      mysql: mysql
      mongodb: mongo-openapi
      redis: redis
      nacos: nacos-openapi
  discovery:
    registry: nacos                    #微服务的服务发现与注册中心类型 nacos,consul,默认是 nacos
    callType: x-form                     #微服务调用参数模式 x-form,json,restful 三种模式可选
  log:
    req: ApiGateRequestLog
  logger:
    level: debug
    out: console,file
    file: /opt/logs/api-gate
  xlang:
    default: zh-cn
    appName: mgate
gate:
  config: mgate-openapi		#网关具体配置文件前缀，放在nacos中，文件名后面会加上-test.yml
```

## 网关配置文件模板及说明

```yaml
api:
  authorization:		#网关授权认证
    need: true		#是否需要认证，不需要为false
    service: openapi-gate		#调用授权认证接口的微服务名
    uri: /api/v1/ecs/operation/check/sign		#授权认证接口地址
    method: GET		#接口方法
    params:		#参数描述，key为验证接口请求中的form参数名，值为http请求header中的参数名
      partnerId: partnerId
      timestamp: timestamp
      authorization: Authorization
    response:		
      data: valid		#响应结果存放在data中的哪个字段，该字段必须是bool型返回值
  gates: #要暴露的服务与接口清单
    openapi-gate:		#微服务名称
      service: openapi-gate		#准确的微服务名
      prefix: /front-gate    #本服务所有接口对外都加上这个前缀
      allow: #暴露的接口前缀或模糊匹配定义
        - path: /api/v1/cdn     #如果全部接口暴露，只需要配一个path，值为/即可
        - path: /api/v1/ecs
      block: #不转发的接口前缀
        - path: /api/v1/ecs/disk    #隐藏掉部分接口
        - path: /api/v1/cdn/domain/**/stop		#模糊匹配
          method: PUT		#同时匹配接口方法
      unauthorized:		#免认证接口匹配定义
        - path: /api/v1/cdn/traffic
          method: GET
    x-lang:
      service: x-lang
      prefix: /api/v2/xlang    #转出接口前缀
      allow: #暴露的接口前缀
        - path: /lang/string/js/list
        - path: /lang/string/app/version
      block:     #不转发的接口前缀
      unauthorized:
        - path: /lang/string/app/version
          method: POST
        - path: /lang/string/js/list
          method: POST
    openapi-admin:
      service: openapi-admin
      prefix: /api/v2/openapi-admin    #转出接口前缀
      allow: #暴露的接口前缀
        - path: /
      block: #不转发的接口前缀
        - path: /admin/v1/lang/list    #隐藏掉部分接口
        - path: /admin/v1/lang/version
      unauthorized:
        - path: /admin/v1/login/
          method: POST
  swagger:	#动态生成swagger文档配置
    show: true  #是否自动生成swagger文档
    description: 测试openapi通用网关	#文档描述
    title: openapi通用网关	#文档标题
    version: api-gate(for openapi test v1.0.0)	#文档版本
    tags:   #分组标签映射
      openapi-gate:		#微服务名
        CDN-其他: 前台-CDN-其他		#分组标签映射，key为原标签，value为映射后标签
        CDN-域名: 前台-CDN-域名
      x-lang:
        语言: I18N
      openapi-admin:
        镜像管理: 后台-镜像管理
        实例: 后台-实例管理
        登录: 后台-登录
```

