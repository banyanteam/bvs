package zkservice

import (
	"time"
	"testing"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

//https://www.jianshu.com/p/e520dea04af6

// func ZkStateString(s *zk.Stat) string {
//     return fmt.Sprintf("Czxid:%d, Mzxid: %d, Ctime: %d, Mtime: %d, Version: %d, Cversion: %d, Aversion: %d, EphemeralOwner: %d, DataLength: %d, NumChildren: %d, Pzxid: %d",
//         s.Czxid, s.Mzxid, s.Ctime, s.Mtime, s.Version, s.Cversion, s.Aversion, s.EphemeralOwner, s.DataLength, s.NumChildren, s.Pzxid)
// }

func callback(event zk.Event) {
    fmt.Println(">>>>>>>>>>>>>>>>>>>")
    fmt.Println(">>path:", event.Path)
    fmt.Println(">>type:", event.Type.String())
    fmt.Println(">>state:", event.State.String())
    fmt.Println("<<<<<<<<<<<<<<<<<<<")
}


func ZkConnect(hosts []string) *zk.Conn{

    //option := zk.WithEventCallback(callback)
    conn, connEvent, err := zk.Connect(hosts, time.Second*5)
     if err != nil {
         fmt.Println(err)
		 if conn != nil {
			 conn.Close()
		 }
		return nil
	}
	
	var state zk.State = zk.StateConnecting
 	for state != zk.StateConnected {
        select {
	        case ch_event := <-connEvent:
			{
				state = ch_event.State
				fmt.Println("path:", ch_event.Path)
                fmt.Println("type:", ch_event.Type.String())
                fmt.Println("state:", ch_event.State.String())
			}
		}
	}

	return conn
}

func ZkChildWatchTest() {
    fmt.Println("ZkChildWatchTest")
	
	var hosts = []string{"localhost:2181"}
	conn := ZkConnect(hosts)
    defer conn.Close()

    // try create root path
    var root_path = "/avideo/srs"

    // check root path exist
    exist, _, err := conn.Exists(root_path)
    if err != nil {
        fmt.Println(err)
        return
    }

    if !exist {
        fmt.Printf("try create root path: %s\n", root_path)
        var acls = zk.WorldACL(zk.PermAll)
        p, err := conn.Create(root_path, []byte("root_value"), 0, acls)
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Printf("root_path: %s create\n", p)
    }

    // try create child node
    // cur_time := time.Now().Unix()
    // ch_path := fmt.Sprintf("%s/ch_%d", root_path, cur_time)
    // var acls = zk.WorldACL(zk.PermAll)
    // p, err := conn.Create(ch_path, []byte("ch_value"), zk.FlagEphemeral, acls)
    // if err != nil {
    //     fmt.Println(err)
    //     return
    // }
    // fmt.Printf("ch_path: %s create\n", p)

    // watch the child events
	var origin_path = "/avideo/srs/load"
    children, s, child_ch, err := conn.ChildrenW(origin_path)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("origin_path[%s] child_count[%d]\n", origin_path, len(children))
    for idx, ch := range children {
        fmt.Printf("%d, %s\n", idx, ch)
    }

    fmt.Printf("watch children result state[%s]\n", ZkStateString(s))

    for {
        select {
        case ch_event := <-child_ch:
            {
                fmt.Println("path:", ch_event.Path)
                fmt.Println("type:", ch_event.Type.String())
                fmt.Println("state:", ch_event.State.String())

                if ch_event.Type == zk.EventNodeCreated {
                    fmt.Printf("has node[%s] detete", ch_event.Path)
                } else if ch_event.Type == zk.EventNodeDeleted {
                    fmt.Printf("has new node[%s] create", ch_event.Path)
                } else if ch_event.Type == zk.EventNodeDataChanged {
                    fmt.Printf("has node[%s] data changed", ch_event.Path)
                } else if ch_event.Type == zk.EventNodeChildrenChanged {
                    children, s, child_ch, err = conn.ChildrenW(origin_path)
				    if err != nil {
				        fmt.Println(err)
				        return
				    }
				
				    fmt.Printf("origin_path[%s] child_count[%d]\n", origin_path, len(children))
				    for idx, ch := range children {
				        fmt.Printf("%d, %s\n", idx, ch)
						child_path := origin_path + "/" + ch
						v, s, err := conn.Get(child_path)
					    if err != nil {
					        fmt.Println(err)
					        return
					    }
					
					    fmt.Printf("value of path[%s]=[%s].\n", child_path, v)
					    fmt.Printf("state: %s\n", ZkStateString(s))
				    }

                }
            }
        }
		
        time.Sleep(time.Millisecond * 10)
    }
}

func ZkNodeManagerTest(){
	var hosts = []string{"localhost:2181"}
	manager := NewZkNodeManager(hosts)
	
	go manager.Start()
	
	for i := 0; i <= 10; i++ {
		time.Sleep(time.Second * 10)
		
		originNode := &StreamingNode{
			StreamType: STREAM_TYPE_RTMP,
		}
		err := manager.GetMinPlayloadOriginSNode(originNode)
	    if err != nil {
	        fmt.Println(err)
	        continue
	    }
		fmt.Printf("originNode LocalHost %s, playload %d\n", 
			originNode.LocalHost, originNode.Load)
	
		edgeNode := &StreamingNode{
			StreamType: STREAM_TYPE_RTMP,
		}
	
		err = manager.GetMinPlayloadEdgeSNode(originNode, edgeNode)
	    if err != nil {
	        fmt.Println(err)
	        continue
	    }
		fmt.Printf("edgeNode LocalHost %s, playload %d\n", 
			edgeNode.LocalHost, edgeNode.Load)
		
	}
}

func TestZk(t *testing.T){
	//ZkChildWatchTest()
	ZkNodeManagerTest()
	t.Log("ZkChildWatchTest")
}

