package routers

import (
	"strconv"
	"strings"
	"fmt"
	"net/url"
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
	log.Println("body: %s", string(data))

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
	
	responseByte, _ := json.Marshal(response)
	log.Println("response body: ", string(responseByte))
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
	
	bytesMsgParams, _ := json.Marshal(request.MsgContent.MsgParams)
	requestParams := KickoffStreamRequestParams{}
	err := json.Unmarshal(bytesMsgParams, &requestParams)
	if err != nil {
		log.Println("Unmarshal ", err)
		response.MsgRetCode = MSG_RETCODE_PARA_ERROR
		response.MsgRetMessage = err.Error()
		return
	}
	
	streamUrl, err := url.Parse(requestParams.TcUrl)
	if err != nil {
		log.Println("url Parse ", err)
		response.MsgRetCode = MSG_RETCODE_PARA_ERROR
		response.MsgRetMessage = err.Error()
		return
	}

	streamNode := zkservice.StreamingNode{
			LocalHost : streamUrl.Host,
		}
	manager := zkservice.NodeManager
	err = manager.GetStreamNode(&streamNode)
	if err != nil {
		log.Println("GetStreamNode ", err)
		response.MsgRetCode = MSG_RETCODE_INNER_ERROR
		response.MsgRetMessage = err.Error()
		return
	}
	
	var clientPort int = 19850
	port, _ := strconv.Atoi(streamUrl.Port())
	clientPort = port + 500
	if port < 10000 {
		clientPort = port + 50
	}
	hostaddr := strings.Split(streamUrl.Host, ":")
	//post message to stream node
	client := &http.Client{}
	clientRequestUrl := fmt.Sprintf("http://%s:%d/api/v1/control", hostaddr[0], clientPort)
	clientRequestBody := ClientKickoffStreamRequestParams {
			Control : "kick",
			TcUrl : requestParams.TcUrl,
			StreamName : requestParams.StreamName,
		}
	clientRequestBodyByte, _ := json.Marshal(clientRequestBody)
	clientRequest, _ := http.NewRequest(http.MethodPost, clientRequestUrl, 
			strings.NewReader(string(clientRequestBodyByte)) )
	clientResponse, err := client.Do(clientRequest)
	clientResponseBody := make([]byte, 0)
	if err == nil {
		if clientResponse.StatusCode == http.StatusOK {
			clientResponseBody, err = ioutil.ReadAll(clientResponse.Body)
		}
	}

	log.Printf("client request url %s, response %s\n", clientRequestUrl, string(clientResponseBody))
	if err != nil {
		log.Println("client request ", err)
		response.MsgRetCode = MSG_RETCODE_INNER_ERROR
		response.MsgRetMessage = err.Error()
	} else {
		response.MsgRetCode = MSG_RETCODE_OK
		response.MsgRetMessage = "success"
	}

	response.MsgData.MsgContent = KickoffStreamReponseParams {
		MsgType : MSG_KICK_OFF_STREAM_RSP,
		Status : string(clientResponseBody),
	}

	return 
}
