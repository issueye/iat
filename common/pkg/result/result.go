package result

import (
	"encoding/json"
	"net/http"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(data interface{}) *Result {
	return &Result{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

func Fail(msg string) *Result {
	return &Result{
		Code: 500,
		Msg:  msg,
		Data: nil,
	}
}

func Error(msg string) *Result {
	return Fail(msg)
}

func (r *Result) JSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(r)
}
