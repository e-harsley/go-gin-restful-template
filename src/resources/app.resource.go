package resources

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	errorhandler "github.com/e-harsley/go-gin-restful-template/src/errorHandler"
	"github.com/e-harsley/go-gin-restful-template/src/services"
	"github.com/e-harsley/go-gin-restful-template/src/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func convertToCamelCase(s string) string {
	tmpl, err := template.New("camelCase").Parse("{{.}}")
	if err != nil {
		panic(err)
	}

	sd := strings.Fields(s)

	st := ""
	for _, x := range sd {
		st = st + strings.Title(x)
	}

	fmt.Println("i am", st)

	camel := strings.ReplaceAll(st, "-", "")
	camel = strings.Title(camel)
	var b bytes.Buffer
	err = tmpl.Execute(&b, camel)
	if err != nil {
		panic(err)
	}
	return b.String()
}

type Pagination struct {
	Limit      int         `json:"limit,omitempty"`
	Page       int         `json:"page,omitempty"`
	Sort       string      `json:"sort,omitempty"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}

func paginate(value interface{}, pagination *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)
	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

type BaseResource struct {
	Resource
	FrontResource interface{}
	Service       services.BaseService
	Serializers   map[string]interface{}
	LimitedQuery  func(*gorm.DB) *gorm.DB
	LimitGet      func(ctx *gin.Context, data map[string]interface{}) (map[string]interface{}, error)
}

func (br *BaseResource) Query() *gorm.DB {
	return br.Service.Db
}

func (br *BaseResource) Get(c *gin.Context) {

	obj_id := c.Param("id")
	fmt.Println("i am object id", obj_id)
	if obj_id == "" {

		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
		page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		sort := c.DefaultQuery("sort", "Id desc")

		p := Pagination{Limit: int(limit), Page: int(page), Sort: sort}

		base_query := br.Query()
		limited_query := br.LimitedQuery(base_query).Scopes(paginate(&br.Service.Model, &p, br.Query())).Preload(clause.Associations).Find(&br.Service.Model)
		var result []map[string]interface{}

		if limited_query.Error != nil {
			errorhandler.BadRequestHandler(c, "failed to fetch items")
			return
		}
		limited_query.Scan(&result)
		fmt.Println("i am response", result)
		if bindTo, ok := br.Serializers["Response"]; ok {
			var responseArray []interface{}
			for _, result := range result {
				jsonBytes, _ := json.Marshal(result)
				structVariable := reflect.New(reflect.TypeOf(bindTo).Elem()).Interface()
				_ = json.Unmarshal(jsonBytes, &structVariable)
				responseArray = append(responseArray, structVariable)
			}
			p.Data = responseArray
			c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": p})
			return
		}

	}

	resp := reflect.ValueOf(br.FrontResource).MethodByName("Fetch").Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(obj_id)})
	res, res_err := resp[0].Interface(), resp[1].Interface()

	if res_err != nil {
		errorhandler.BadRequestHandler(c, res_err)
		return
	}

	respon, err := br.LimitGet(c, res.(map[string]interface{}))

	if res_err != nil {
		errorhandler.BadRequestHandler(c, err.Error())
		return
	}

	if bindTo, ok := br.Serializers["Response"]; ok {
		jsonBytes, _ := json.Marshal(respon)
		structVariable := reflect.New(reflect.TypeOf(bindTo).Elem()).Interface()
		_ = json.Unmarshal(jsonBytes, &structVariable)
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": structVariable})
		return
	}
}

func (br *BaseResource) Fetch(c *gin.Context, obj_id string) (map[string]interface{}, error) {
	id, _ := strconv.ParseUint(obj_id, 10, 64)
	return br.Service.Get(uint(id))

}

func (br *BaseResource) Post(c *gin.Context) {
	obj_id := c.Param("id")
	action := c.Param("action")

	if obj_id == "" && action == "" {
		if serializer, ok := br.Serializers["Request"]; ok {
			jsonRequest := reflect.New(reflect.TypeOf(serializer).Elem()).Interface()

			err := c.ShouldBindJSON(&jsonRequest)
			if err != nil {
				var val validator.ValidationErrors
				if errors.As(err, &val) {
					errorhandler.ConflictHandler(c, utils.ValidationErrorResponse(val))
					return
				}
				errorhandler.ConflictHandler(c, err.Error())
				return
			}
			fmt.Println(jsonRequest)

			// convert to map[string]interface{}
			jsonBytes, _ := json.Marshal(jsonRequest)
			var result map[string]interface{}
			json.Unmarshal(jsonBytes, &result)

			fmt.Println(result)
			resp := reflect.ValueOf(br.FrontResource).MethodByName("Save").Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(result)})
			res, res_err := resp[0].Interface(), resp[1].Interface()

			if res_err != nil {
				errorhandler.BadRequestHandler(c, res_err)
				return
			}
			if bindTo, ok := br.Serializers["Response"]; ok {
				jsonBytes, _ := json.Marshal(res)
				structVariable := reflect.New(reflect.TypeOf(bindTo).Elem()).Interface()
				_ = json.Unmarshal(jsonBytes, &structVariable)
				c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": structVariable})
				return
			}
			errorhandler.ConflictHandler(c, "response serializer not found")
			return
		}
		errorhandler.ConflictHandler(c, "request serializer not found")
	}
	if obj_id == "" && action != "" {
		method := reflect.ValueOf(br.FrontResource).MethodByName(convertToCamelCase(action))
		if method.IsValid() {
			if serializer, ok := br.Serializers[convertToCamelCase(action)]; ok {
				jsonRequest := reflect.New(reflect.TypeOf(serializer).Elem()).Interface()

				err := c.ShouldBindJSON(&jsonRequest)
				if err != nil {
					var val validator.ValidationErrors
					if errors.As(err, &val) {
						errorhandler.ConflictHandler(c, utils.ValidationErrorResponse(val))
						return
					}
					errorhandler.ConflictHandler(c, err.Error())
					return
				}
				fmt.Println(jsonRequest)

				// convert to map[string]interface{}
				jsonBytes, _ := json.Marshal(jsonRequest)
				var result map[string]interface{}
				json.Unmarshal(jsonBytes, &result)

				fmt.Println(result)
				resp := reflect.ValueOf(br.FrontResource).MethodByName(convertToCamelCase(action)).Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(result)})
				res, res_err := resp[0].Interface(), resp[1].Interface()

				if res_err != nil {
					errorhandler.BadRequestHandler(c, res_err)
					return
				}
				if bindTo, ok := br.Serializers[convertToCamelCase(action)+"Response"]; ok {
					jsonBytes, _ := json.Marshal(res)
					structVariable := reflect.New(reflect.TypeOf(bindTo).Elem()).Interface()
					_ = json.Unmarshal(jsonBytes, &structVariable)
					c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": structVariable})
					return
				} else if resJ, ok := br.Serializers["Response"]; ok {
					jsonBytes, _ := json.Marshal(res)
					structVariable := reflect.New(reflect.TypeOf(resJ).Elem()).Interface()
					_ = json.Unmarshal(jsonBytes, &structVariable)
					c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": structVariable})
					return
				}
				errorhandler.ConflictHandler(c, "response serializer not found")
				return
			}
			errorhandler.ConflictHandler(c, "request serializer not found")
		}
		errorhandler.NotFoundHandler(c, "not found")
		return
	}
	if obj_id != "" && action != "" {
		fmt.Println(action)
		resp := reflect.ValueOf(br.FrontResource).MethodByName("Fetch").Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(obj_id)})
		res, res_err := resp[0].Interface(), resp[1].Interface()

		if res_err != nil {
			errorhandler.BadRequestHandler(c, res_err)
			return
		}

		_, err := br.LimitGet(c, res.(map[string]interface{}))

		if res_err != nil {
			errorhandler.BadRequestHandler(c, err.Error())
			return
		}
		fmt.Println(convertToCamelCase(action))
		method := reflect.ValueOf(br.FrontResource).MethodByName(convertToCamelCase(action))
		if method.IsValid() {
			fmt.Println("i go here")
			if serializer, ok := br.Serializers[convertToCamelCase(action)]; ok {
				jsonRequest := reflect.New(reflect.TypeOf(serializer).Elem()).Interface()

				err := c.ShouldBindJSON(&jsonRequest)
				if err != nil {
					var val validator.ValidationErrors
					if errors.As(err, &val) {
						errorhandler.ConflictHandler(c, utils.ValidationErrorResponse(val))
						return
					}
					errorhandler.ConflictHandler(c, err.Error())
					return
				}
				fmt.Println(jsonRequest)

				// convert to map[string]interface{}
				jsonBytes, _ := json.Marshal(jsonRequest)
				var result map[string]interface{}
				json.Unmarshal(jsonBytes, &result)

				fmt.Println(result)
				resp := reflect.ValueOf(br.FrontResource).MethodByName(convertToCamelCase(action)).Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(obj_id), reflect.ValueOf(result)})
				res, res_err := resp[0].Interface(), resp[1].Interface()

				if res_err != nil {
					errorhandler.BadRequestHandler(c, res_err)
					return
				}
				if bindTo, ok := br.Serializers[convertToCamelCase(action)+"Response"]; ok {
					jsonBytes, _ := json.Marshal(res)
					structVariable := reflect.New(reflect.TypeOf(bindTo).Elem()).Interface()
					_ = json.Unmarshal(jsonBytes, &structVariable)
					c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": structVariable})
					return
				} else if resJ, ok := br.Serializers["Response"]; ok {
					jsonBytes, _ := json.Marshal(res)
					structVariable := reflect.New(reflect.TypeOf(resJ).Elem()).Interface()
					_ = json.Unmarshal(jsonBytes, &structVariable)
					c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": structVariable})
					return
				}
				errorhandler.ConflictHandler(c, "response serializer not found")
				return
			}
			errorhandler.ConflictHandler(c, "request serializer not found")
			return
		}
		errorhandler.NotFoundHandler(c, "not found")
		return
	}
	errorhandler.NotFoundHandler(c, "not found")

}

func (br *BaseResource) Save(c *gin.Context, data map[string]interface{}) (map[string]interface{}, error) {
	return br.Service.Create(data)
}

func (br *BaseResource) Update(c *gin.Context, obj_id string, data map[string]interface{}) (map[string]interface{}, error) {
	id, _ := strconv.ParseUint(obj_id, 10, 64)
	return br.Service.Update(uint(id), data)
}

func (br *BaseResource) Put(c *gin.Context) {
	obj_id := c.Param("id")
	resp := reflect.ValueOf(br.FrontResource).MethodByName("Fetch").Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(obj_id)})
	res, res_err := resp[0].Interface(), resp[1].Interface()

	if res_err != nil {
		errorhandler.BadRequestHandler(c, res_err)
		return
	}

	_, err := br.LimitGet(c, res.(map[string]interface{}))

	if res_err != nil {
		errorhandler.BadRequestHandler(c, err.Error())
		return
	}

	if serializer, ok := br.Serializers["Request"]; ok {
		jsonRequest := reflect.New(reflect.TypeOf(serializer).Elem()).Interface()

		err := c.ShouldBindJSON(&jsonRequest)
		if err != nil {
			var val validator.ValidationErrors
			if errors.As(err, &val) {
				errorhandler.ConflictHandler(c, utils.ValidationErrorResponse(val))
				return
			}
			errorhandler.ConflictHandler(c, err.Error())
			return
		}
		fmt.Println(jsonRequest)

		// convert to map[string]interface{}
		jsonBytes, _ := json.Marshal(jsonRequest)
		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		fmt.Println(result)
		resp := reflect.ValueOf(br.FrontResource).MethodByName("Update").Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(obj_id), reflect.ValueOf(result)})
		res, res_err := resp[0].Interface(), resp[1].Interface()

		if res_err != nil {
			errorhandler.BadRequestHandler(c, res_err)
			return
		}
		if bindTo, ok := br.Serializers["Response"]; ok {
			jsonBytes, _ := json.Marshal(res)
			structVariable := reflect.New(reflect.TypeOf(bindTo).Elem()).Interface()
			_ = json.Unmarshal(jsonBytes, &structVariable)
			c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": structVariable})
			return
		}
		errorhandler.ConflictHandler(c, "response serializer not found")
		return
	}
	errorhandler.ConflictHandler(c, "request serializer not found")

}

func (br *BaseResource) Delete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "This is a Delete method for the resource"})
}

func (br *BaseResource) Initialize(resource interface{}, service services.BaseService, serializer map[string]interface{}, limitter func(*gorm.DB) *gorm.DB, LimitGet func(ctx *gin.Context, data map[string]interface{}) (map[string]interface{}, error)) {
	br.Service = service
	br.Serializers = serializer
	br.FrontResource = resource
	br.LimitGet = LimitGet
	br.LimitedQuery = limitter
}

// func NewResource[T comparable](service services.BaseService[T]) *BaseResource {
// 	return &BaseResource{
// 		Service: service,
// 	}

// }

// func (sm *TestResource) Get(c *gin.Context) {

// 	sm.Service.
// 		c.JSON(http.StatusOK, gin.H{"message": "This is a test get method for the resource"})
// }
