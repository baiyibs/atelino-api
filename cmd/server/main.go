package main

//	@title			Atelino API
//	@version		1.0.1
//	@description	Atelino 后端 API 文档
//	@termsOfService	http://swagger.io/terms/

//	@host		localhost:8080
//	@BasePath	/api/

//	@securityDefinitions.basic	BearerAuth

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
import (
	"atelino/internal/app"
	_ "atelino/pkg/docs"
)

func main() {
	app.Run()
}
