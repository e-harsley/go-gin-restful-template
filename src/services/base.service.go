package services

import "gorm.io/gorm"

type ServiceFactory struct {
	DB *gorm.DB
}

func (sf ServiceFactory) CreateService(model interface{}) *BaseService {
	return &BaseService{
		Db:    sf.DB,
		Model: model,
	}
}

var FactoryService ServiceFactory
