package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DatabaseConnection(host, dbName, user, password string, port int) (*gorm.DB, error) {

	sqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	return gorm.Open(postgres.Open(sqlInfo), &gorm.Config{})
}
