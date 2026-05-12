package main

import (
	"atelino/internal/app"

	//	@title			Atelino API
	//	@version		1.1.1
	//	@description	Atelino 后端 API 文档

	//	@host						localhost:8080
	//	@BasePath					/
	//	@schemes					http https
	//	@securityDefinitions.apikey	BearerAuth
	//	@in							header
	//	@name						Authorization

	//	@externalDocs.description	OpenAPI
	//	@externalDocs.url			https://swagger.io/resources/open-api/

	_ "atelino/pkg/docs"
)

func main() {
	app.Run()
}
