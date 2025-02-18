package main

import (
	"awesomeProject3/database"
	"awesomeProject3/routes"
	"log"
)

func main() {
	database.InitDB()
	r := routes.SetupRouter()
	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
