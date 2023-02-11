package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

func DecodeMapToJson(data map[string]interface{}, jsonResponse interface{}) (interface{}, error) {
	config := &mapstructure.DecoderConfig{
		Result:  &jsonResponse,
		TagName: "mapstructure",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return jsonResponse, err
	}
	err = decoder.Decode(data)
	if err != nil {
		return jsonResponse, err
	}
	return jsonResponse, nil
}

func ValidationErrorResponse(val validator.ValidationErrors) map[string]string {

	errs := make(map[string]string)

	for _, f := range val {
		err := f.ActualTag()
		if f.Param() != "" {
			err = fmt.Sprintf("%s=%s", err, f.Param())
		}
		errs[strings.ToLower(f.Field())] = err
	}
	return errs
}
