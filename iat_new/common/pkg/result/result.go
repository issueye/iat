package result

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
