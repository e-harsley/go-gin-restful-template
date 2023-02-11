package services

import (
	"fmt"
	"strconv"

	specification "github.com/e-harsley/go-gin-restful-template/src/specifications"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseService struct {
	Model interface{}
	Db    *gorm.DB
}

func (b *BaseService) Get(id uint) (map[string]interface{}, error) {

	var result map[string]interface{}

	// recId := strconv.Itoa(int(id))

	dbResponse := b.Db.Model(&b.Model).Preload(clause.Associations).First(&b.Model, id)

	err := dbResponse.Error

	if err != nil {
		return result, err
	}

	dbResponse.Scan(&result)

	return result, nil
}

func (b *BaseService) FindOne(specification specification.Specification) (map[string]interface{}, error) {

	var result map[string]interface{}

	dbResponse := b.Db.Where(specification.GetQuery(), specification.GetValues()...).First(&b.Model)

	err := dbResponse.Error

	if err != nil {
		return result, err
	}

	dbResponse.Scan(&result)

	return result, nil
}

func (b *BaseService) Create(data map[string]interface{}) (map[string]interface{}, error) {

	var result map[string]interface{}

	err := b.Db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&b.Model).Create(data).Error
		if err != nil {
			return err
		}
		res := tx.Last(&b.Model)

		if res.Error != nil {
			return res.Error
		}
		res.Scan(&result)
		return nil
	})

	if err != nil {
		return result, err
	}

	return result, nil
}

func (b *BaseService) Update(id uint, data map[string]interface{}) (map[string]interface{}, error) {

	var result map[string]interface{}

	singleRecord, err := b.Get(id)

	if err != nil {
		return result, err
	}
	fmt.Println("i got here", singleRecord)

	recId := strconv.Itoa(int(singleRecord["id"].(int64)))

	dbResponse := b.Db.Model(&b.Model).Where("id = " + recId).Updates(data)

	err = dbResponse.Error

	if err != nil {
		return result, err
	}

	dbResponse.Scan(&result)

	return result, nil
}

func (b *BaseService) Delete(id uint) (map[string]interface{}, error) {

	var result map[string]interface{}

	singleRecord, err := b.Get(id)

	if err != nil {
		return result, err
	}

	dbResponse := b.Db.Model(&b.Model).Delete(b.Model, singleRecord["id"].(uint))

	err = dbResponse.Error

	if err != nil {
		return result, err
	}

	dbResponse.Scan(&result)

	return result, nil
}
