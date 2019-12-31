package routers

import(
	"time"
)

const (
	MSG_GET_STREAM_SERVER_ADDR_REQ = "MSG_GET_STREAM_SERVER_ADDR_REQ"
	MSG_GET_STREAM_SERVER_ADDR_RSP = "MSG_GET_STREAM_SERVER_ADDR_RSP"
	MSG_KICK_OFF_STREAM_REQ = "MSG_KICK_OFF_STREAM_REQ"
	MSG_KICK_OFF_STREAM_RSP = "MSG_KICK_OFF_STREAM_RSP"
)

const (
	MSG_RETCODE_OK = 1
	MSG_RETCODE_INNER_ERROR = -100
	MSG_RETCODE_PARA_ERROR = -101
)

type MsgRequestContent struct {
	MsgType string      `json:"request"`
	MsgParams interface{}     `json:"requestParams"`
}

type MessageRequest struct {
	MsgSession int64      `json:"msgSession"`
	MsgSequence int       `json:"msgSequence"`
	MsgTimeStamp int64    `json:"msgTimeStamp"`
	MsgCategory string    `json:"msgCategory"`
	MsgContent  MsgRequestContent   `json:"msgContent"`
}

type MsgData struct {
	MsgSession int64      `json:"msgSession"`
	MsgSequence int       `json:"msgSequence"`
	MsgTimeStamp int64    `json:"msgTimeStamp"`
	MsgCategory string    `json:"msgCategory"`
	MsgContent  interface{}    `json:"msgContent"`
}

type MessageResponse struct {
	MsgRetCode int64      `json:"code"`
	MsgRetMessage string     `json:"message"`
	MsgData MsgData       `json:"data"`
}

func NewMessageResponse(request *MessageRequest) *MessageResponse{

	response := &MessageResponse{
		MsgData : MsgData{
				MsgSession : request.MsgSession,
				MsgSequence : request.MsgSequence,
				MsgCategory : request.MsgCategory,
				MsgTimeStamp : time.Now().Unix(),
			},
	}
	
	return response
}

/*
MSG_GET_STREAM_SERVER_ADDR_REQ 

"requestParams" : {
	"origin" : {
		"streamType" : 0,
		"assigned" : 0,
        "localHost" : "172.16.3.161:1935",
        "publicHost" : "117.139.13.231:1935",
	},
    "edge" : {
        "streamType" : 0,
	}
}
*/
type StreamingNodeParam struct {
	StreamType int       `json:"streamType"`
	Assigned int    `json:"assigned"`
	LocalHost string    `json:"localHost"`
	PublicHost string    `json:"publicHost"`
}

type StreanServerAddrRequestParams struct {
	Origin StreamingNodeParam `json:"origin"`
	Edge StreamingNodeParam `json:"edge"`
}

type StreanServerAddrReponseParams struct {
	MsgType string `json:"response"`
	Origin StreamingNodeParam `json:"origin"`
	Edge StreamingNodeParam `json:"edge"`
}
