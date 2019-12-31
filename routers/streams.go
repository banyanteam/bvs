package routers

import (
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/bvs/zkservice"
)

func (h *APIHandler) StreamService(c *gin.Context) {
	log.Println("request: ", c.Request.URL)

	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("body: %s\n", string(data))

	request := &MessageRequest{}
	err = json.Unmarshal(data, request)
	if err != nil {
		log.Println("Unmarshal request error ", err)
		return
	}
	
	response := NewMessageResponse(request)
	if request.MsgContent.MsgType == MSG_GET_STREAM_SERVER_ADDR_REQ {
		h.GetStreanServerAddrReq(request, response)
	} else if request.MsgContent.MsgType == MSG_KICK_OFF_STREAM_REQ {
		h.KickStream(request, response)
	}
	
	c.IndentedJSON(http.StatusOK, response)	

	return
}

func (h *APIHandler) GetStreanServerAddrReq(request *MessageRequest, response *MessageResponse) {

	bytesMsgParams, _ := json.Marshal(request.MsgContent.MsgParams)
	requestParams := StreanServerAddrRequestParams{}
	err := json.Unmarshal(bytesMsgParams, &requestParams)
	if err != nil {
		response.MsgRetCode = MSG_RETCODE_PARA_ERROR
		response.MsgRetMessage = err.Error()
		return
	}

	manager := zkservice.NodeManager
	originNode := zkservice.StreamingNode{}
	if requestParams.Origin.Assigned == 0 {
		err = manager.GetMinPlayloadOriginSNode(&originNode)
		if err != nil {
			response.MsgRetCode = MSG_RETCODE_INNER_ERROR
			response.MsgRetMessage = err.Error()
			return
		}
	} else {
		originNode.StreamType = requestParams.Origin.StreamType
		originNode.LocalHost = requestParams.Origin.LocalHost
	}
	
	edgeNode := zkservice.StreamingNode{}
	if edgeNode.StreamType == zkservice.STREAM_TYPE_HTTPFLV {
		//for http+flv use origin node as edge node
		edgeNode = originNode
	} else {
		err = manager.GetMinPlayloadEdgeSNode(&originNode, &edgeNode)
		if err != nil {
			if edgeNode.StreamType == zkservice.STREAM_TYPE_RTMP {
				// can use origin node as edge node
				edgeNode = originNode
			} else {
				response.MsgRetCode = MSG_RETCODE_INNER_ERROR
				response.MsgRetMessage = err.Error()
				return
			}
		}
	}
	
	response.MsgRetCode = MSG_RETCODE_OK
	response.MsgRetMessage = "success"
	
	response.MsgData.MsgContent = StreanServerAddrReponseParams {
		MsgType : MSG_GET_STREAM_SERVER_ADDR_RSP,
		Origin : StreamingNodeParam{
				StreamType : originNode.StreamType,
				LocalHost : originNode.LocalHost,
				PublicHost : originNode.PublicHost,
			},
		Edge : StreamingNodeParam{
				StreamType : edgeNode.StreamType,
				LocalHost : edgeNode.LocalHost,
				PublicHost : edgeNode.PublicHost,
			},
	}
	
	return
}

func (h *APIHandler) KickStream(request *MessageRequest, response *MessageResponse) {
	
}
