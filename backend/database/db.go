package database

import (
	"awesomeProject3/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB


func InitDB() *gorm.DB {
	dsn := "host=127.0.0.1 port=5433 dbname=warehouse user=postgres password=postgres connect_timeout=10 sslmode=prefer"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	log.Println("Подключение к базе данных установлено.")

	if err := DB.AutoMigrate(&models.User{}, &models.Product{}, &models.Invoice{}, &models.InvoiceProduct{}, &models.Storage{}); err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}

	return DB
}
