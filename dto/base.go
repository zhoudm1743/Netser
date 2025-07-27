package dto

import (
	"encoding/json"
)

type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// args [1]data, [2]message [3]code
func Success(args ...any) string {
	var data any
	var message string
	var code int

	if len(args) > 0 {
		data = args[0]
	} else {
		data = nil
	}
	if len(args) > 1 {
		message = args[1].(string)
	} else {
		message = "success"
	}
	if len(args) > 2 {
		code = args[2].(int)
	} else {
		code = 1
	}

	resp := BaseResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
	json, _ := json.Marshal(resp)
	return string(json)
}

func Error(args ...any) string {
	var message string
	var code int

	if len(args) > 0 {
		message = args[0].(string)
	} else {
		message = "error"
	}
	if len(args) > 1 {
		code = args[1].(int)
	} else {
		code = 0
	}
	resp := BaseResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	json, _ := json.Marshal(resp)
	return string(json)
}

type BaseRequest struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

func (b *BaseRequest) Unmarshal(str string) error {
	object := BaseRequest{}
	err := json.Unmarshal([]byte(str), &object)
	if err != nil {
		return err
	}
	*b = object
	return nil
}
