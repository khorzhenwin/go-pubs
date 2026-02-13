package site

import (
	"encore.dev/storage/sqldb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

var siteDB = sqldb.Named("site").Stdlib()

// initService initializes the site service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: siteDB,
	}))
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}
