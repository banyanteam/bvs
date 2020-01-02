
package zkservice

import (
	"errors"
	"time"
	"sync"
	"fmt"
	"log"
	"encoding/json"
	"github.com/samuel/go-zookeeper/zk"
)

const INT_MAX = int(^uint(0) >> 1)

const (
	STREAM_TYPE_RTMP int = iota
	STREAM_TYPE_RTSP
	STREAM_TYPE_RTZ
	STREAM_TYPE_HTTPFLV
	STREAM_TYPE_HLS
)

type StreamingNode struct {
	Mode         int        `json:"mode"`
	PublicHost   string        `json:"public_host"`
	LocalHost    string        `json:"local_host"`
	OriginHost   string        `json:"origin_host"`
	Load	     int        `json:"load"`

	StreamType   int
	Ctime        int64
}

type OriginSNode struct {
	StreamingNode
	edges       map[string]*StreamingNode //localHost <-> StreamingNode
}

type ZkNodeManager struct {

	mLock    	sync.RWMutex
	origins     map[string]*OriginSNode //localHost <-> *OriginSNode

	zkHosts 	[]string
	conn		*zk.Conn
}

var NodeManager *ZkNodeManager

func NewOriginSNode() (origin *OriginSNode) {
	origin = &OriginSNode{
		edges:  make(map[string]*StreamingNode),
	}
	return
}

func NewZkNodeManager(hosts []string) {
	NodeManager = &ZkNodeManager{
		origins: make(map[string]*OriginSNode),
		zkHosts: hosts,
	}
	return
}

func ZkStateString(s *zk.Stat) string {
    return fmt.Sprintf("Czxid:%d, Mzxid: %d, Ctime: %d, Mtime: %d, Version: %d, Cversion: %d, Aversion: %d, EphemeralOwner: %d, DataLength: %d, NumChildren: %d, Pzxid: %d",
        s.Czxid, s.Mzxid, s.Ctime, s.Mtime, s.Version, s.Cversion, s.Aversion, s.EphemeralOwner, s.DataLength, s.NumChildren, s.Pzxid)
}

func (manager *ZkNodeManager) Start(){

	conn := manager.conn
	defer func() {
		if conn != nil{
			conn.Close()
		}
	}()

	err := manager.connectZk()
	if err != nil {
		log.Println("connectZk ", err)
		return
	}
	
	for {
		err = manager.traverseStreamingNode()
		if err != nil {
			log.Println("traverseStreamingNode ", err)
			return
		}
		
		time.Sleep(time.Second * 10)
	}
}


func (manager *ZkNodeManager) GetMinPlayloadOriginSNode(originNode *StreamingNode) error{

	var minPlayload int = INT_MAX
	
	manager.mLock.RLock()
	for _, origin := range manager.origins {
		if originNode.StreamType == origin.StreamType && origin.Load < minPlayload {
			minPlayload = origin.Load
			*originNode = origin.StreamingNode
		}
	}
	
	manager.mLock.RUnlock()
	if minPlayload == INT_MAX {
		log.Printf("can not find origin node, type %d\n", originNode.StreamType)
		return errors.New("can not find origin node")
	}
	return nil
}

func (manager *ZkNodeManager) GetMinPlayloadEdgeSNode(originNode *StreamingNode, edgeNode *StreamingNode) error{

	manager.mLock.RLock()
	origin, ok := manager.origins[originNode.LocalHost]
	if !ok {
		log.Printf("can not find edge node, origin node %s\n", originNode.LocalHost)
		manager.mLock.RUnlock()
		return errors.New("can not find edge node")
	}
	
	var minPlayload int = INT_MAX
	for _, edge := range origin.edges {
		if edge.StreamType == edgeNode.StreamType && edge.Load < minPlayload {
			minPlayload = edge.Load
			*edgeNode = *edge
		}
	}
	
	manager.mLock.RUnlock()
	if minPlayload == INT_MAX {
		log.Printf("can not find edge node, origin node %s, edge node type %d\n", 
			originNode.LocalHost, edgeNode.StreamType)
		return errors.New("can not find edge node")
	}

	return nil
}

func (manager *ZkNodeManager) GetStreamNode(streamNode *StreamingNode) error{

	manager.mLock.RLock()
	for _, origin := range manager.origins {
		if origin.LocalHost == streamNode.LocalHost {
			*streamNode = origin.StreamingNode
			manager.mLock.RUnlock()
			return nil
		}

		for _, edge := range origin.edges {
			if edge.LocalHost == streamNode.LocalHost {
				*streamNode = *edge
				manager.mLock.RUnlock()
				return nil
			}
		}
	}

	log.Printf("can not find stream node, localhost %s\n", streamNode.LocalHost)
	manager.mLock.RUnlock()
	return errors.New("can not find stream node")
}

func (manager *ZkNodeManager) connectZk() error{

	log.Printf("connecting zk %v ... ", manager.zkHosts)
    conn, connEvent, err := zk.Connect(manager.zkHosts, time.Second*5)
	if err != nil {
	   	log.Println(err)
		return err
	}
	
	var state zk.State = zk.StateConnecting
 	for state != zk.StateConnected {
        select {
	        case ch_event := <-connEvent:
			{
				state = ch_event.State
				// fmt.Println("path:", ch_event.Path)
				// fmt.Println("type:", ch_event.Type.String())
				// fmt.Println("state:", ch_event.State.String())
			}
		}
	}
	manager.conn = conn
	log.Println("connect zk successfully!")
	return nil
}


func (manager *ZkNodeManager) traverseStreamingNode() error{
    // try create root path
    var root_path = "/avideo/srs"

	conn := manager.conn
    // check root path exist
    exist, _, err := conn.Exists(root_path)
    if err != nil {
        log.Println(err)
        return err
    }

    if !exist {
        log.Printf("try create root path: %s\n", root_path)
        var acls = zk.WorldACL(zk.PermAll)
        p, err := conn.Create(root_path, []byte("root_value"), 0, acls)
        if err != nil {
            fmt.Println(err)
            return err
        }
        log.Printf("root_path: %s create\n", p)
    }

    // traverse the origin node
	var origin_path = "/avideo/srs/load"
    children, _, _, err := conn.ChildrenW(origin_path)
    if err != nil {
        log.Println(err)
        return err
    }

	tmpOriginSNodes := make(map[string]*OriginSNode)

    // log.Printf("origin_path[%s] child_count[%d]\n", origin_path, len(children))
    for idx, ch := range children {
        // log.Printf("[%d] %s\n", idx, ch)
		child_path := origin_path + "/" + ch
		v, s, err := conn.Get(child_path)
	    if err != nil {
	        log.Println(err)
	        return err
	    }

	    log.Printf("origin node[%d] value of path[%s]=[%s]\n", idx, child_path, v)
	    // log.Printf("state: %s\n", ZkStateString(s))
		
		//get origin node
		originSNode := NewOriginSNode()
		err = json.Unmarshal(v, &originSNode.StreamingNode)
		if err != nil {
			log.Println("Unmarshal error ", err)
			return err
		}
		
		originSNode.Ctime = s.Ctime
		tmpNode, ok := tmpOriginSNodes[originSNode.LocalHost]
		if ok && tmpNode.Ctime > originSNode.Ctime {
			log.Printf("tmpNode is out of date %d vs %d\n", tmpNode.Ctime, originSNode.Ctime)
			continue
		}

		tmpOriginSNodes[originSNode.LocalHost] = originSNode
    }

	// traverse the edge node
	var edge_path = "/avideo/srs/edge"
    children, _, _, err = conn.ChildrenW(edge_path)
    if err != nil {
        log.Println(err)
        return err
    }

    // log.Printf("edge_path[%s] child_count[%d]\n", edge_path, len(children))
    for idx, ch := range children {
        // log.Printf("[%d] %s\n", idx, ch)
		tmpOriginNode, ok := tmpOriginSNodes[ch]
		if !ok {
			log.Printf("edge node[%d] origin node %s does not exist\n", idx, ch)
			continue
		}

		manager.traverseEdgeNodes(edge_path+"/"+ch, STREAM_TYPE_RTMP, tmpOriginNode)
		manager.traverseEdgeNodes(edge_path+"/"+ch, STREAM_TYPE_RTZ, tmpOriginNode)
    }

	manager.mLock.Lock()
	manager.origins = tmpOriginSNodes
	manager.mLock.Unlock()

	return nil
}

func (manager *ZkNodeManager) traverseEdgeNodes(edge_path string, streamType int, tmpOriginNode *OriginSNode) error{

	conn := manager.conn

	var grand_edge_path string
	if streamType == STREAM_TYPE_RTMP {
		grand_edge_path = edge_path + "/srs"
	} else if streamType == STREAM_TYPE_RTZ {
		grand_edge_path = edge_path + "/rtz"
	}

    exist, _, err := conn.Exists(grand_edge_path)
    if err != nil {
        log.Println(err)
        return err
    }

	if exist {
	    grandchildren, _, _, err := conn.ChildrenW(grand_edge_path)
	    if err != nil {
	        log.Println(err)
	        return err
	    }

		tmpEdgeNodes := make(map[string]*StreamingNode)
		for grandidx, grandch := range grandchildren {
			grandchild_path := grand_edge_path + "/" + grandch
			v, s, err := conn.Get(grandchild_path)
		    if err != nil {
		        log.Println(err)
		        return err
		    }
		
		    log.Printf("[%d] value of path[%s]=[%s]\n", grandidx, grandchild_path, v)
		    // log.Printf("state: %s\n", ZkStateString(s))
			
			edgeNode := &StreamingNode{
							StreamType : streamType,
							Ctime : s.Ctime,
						}
			err = json.Unmarshal(v, edgeNode)
			if err != nil {
				log.Println("Unmarshal error ", err)
				return err
			}
			
			tmpNode, ok := tmpEdgeNodes[edgeNode.LocalHost]
			if ok && tmpNode.Ctime > edgeNode.Ctime {
				log.Printf("tmp edge node is out of date %d vs %d\n", tmpNode.Ctime, edgeNode.Ctime)
				continue
			}
	
			tmpEdgeNodes[edgeNode.LocalHost] = edgeNode
		}

		tmpOriginNode.edges = tmpEdgeNodes
	}

	return nil
}


