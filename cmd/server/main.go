package main

//	@title			Atelino API
//	@version		1.0.2
//	@description	Atelino 后端 API 文档

//	@host		localhost:8080
//	@BasePath	/api/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
import (
	"atelino/internal/app"
	_ "atelino/pkg/docs"
)

func main() {
	app.Run()
}
