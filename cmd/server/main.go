package main

//	@title			Atelino API
//	@version		1.0.3
//	@description	Atelino 后端 API 文档

//	@host		localhost:8080
//	@BasePath	/

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
