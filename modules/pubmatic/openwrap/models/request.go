package models

type UnwrapReq struct {
	Adm       string
	BidId     string
	UnwrapCnt int
	Err       error
	RespTime  int
}

type RequestCtx struct {
	//User Agent
	UA             string
	VastUnwrapFlag bool
}
