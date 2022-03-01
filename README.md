# mgate 通用微服务网关

mgate是一款基于Go语言开发的，以Nacos作为注册中心与配置中心的，具备动态添加微服务接口转发配置的通用型微服务网关。

mgate用MongoDB作为API接口配置数据库，用Gin作为Web框架，同时提供动态Swagger文档生成功能。

## 基本功能

- 可以配置网关服务名与端口
- 支持http和https双端口并存
- 支持配置https证书
- 可以方便地从需要导入服务的swagger文档直接导入相应接口映射并自动生成该接口swagger文档
- 无swagger的微服务接口可以手工添加并生成swagger接口文档
- 支持POST和GET两种模式的接口
- 支持application/x-www-form-urlencoded和文件上传的multipart/form-data的content-type接口
- 支持接口访问日志保存到MongoDB
- 支持文本日志文件

## 安装部署

1. **部署Nacos**

2. **部署MongoDB，并创建mgate所需的数据库与授权用户**

3. **在Nacos配置中心中添加一个mongo-mgate-test.yml的配置**

   ```yaml
   go:
     data:
       mongodb:
         uri: mongodb://<username>:<password>@59.56.77.23:2717/<db>
         db: <db>
       mongo_pool:
         min: 2
         max: 20
         idle: 10
         timeout: 300
   ```

   

4. **在Nacos配置中心添加一个nacos-test.yml的配置**

   ```yaml
   go:
     nacos:
       server: xxx.xxx.xxx.xxx
       port: 8848
       clusterName: DEFAULT
       weight: 1
   ```

   

5. **下载本项目**

   ```powershell
   git clone https://github.com/maczh/mgate.git
   ```

   

6. **修改本地配置文件**mgate.yml

   ```yaml
   go:
     application:
       name: mgate
       port: 8088
   #    port_ssl: 8089
   #    cert: certs/gate.pem
   #    key: certs/gate.key
     config:
       server: http://xxx.xxx.xxx.xxx:8848/
       server_type: nacos
       env: test
       type: .yml
       mid: "-"
       used: nacos,mongodb
       prefix:
         mysql: mysql
         mongodb: mongo-mgate
         redis: redis
         rabbitmq: rabbitmq
         nacos: nacos
     log:
       req: mGatewayApiRequestLog
     logger:
       level: debug
       out: console,file
       file: /opt/logs/mgate
   api:
     collection: mgGatewayConfig
   swagger:
     info:
       title: 测试通用API网关
       description: 测试通用网关
       version: "1.0.0(mgate)"
   ```

   

7. **编译**

   ```powershell
   go mod tidy -compat=1.17
   go build
   ```

   

8. **运行**

   ```
   ./mgate
   ```

   

## 管理操作Swagger界面

1. 在浏览器中访问mgate swagger接口文档 http://xxx.xxx.xxx.xxx:8088/docs/index.html
2. 添加网关api接口映射，有两种模式：
   1. 从微服务swagger添加一个网关映射
   2. 直接添加一个网关映射
