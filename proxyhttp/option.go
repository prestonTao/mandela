package proxyhttp

// import (
// 	"mandela/config"
// 	"mandela/core/message_center"
// 	"mandela/core/message_center/flood"
// 	"mandela/core/nodeStore"
// 	"encoding/json"
// )

// /*
// 	获取远程节点web端口信息
// */
// func GetRemoteWebPort(id nodeStore.AddressNet) uint16 {
// 	message, ok := message_center.SendP2pMsgHE(config.MSGID_http_getwebinfo, id, nil)
// 	if ok {
// 		bs := flood.WaitRequest(message_center.CLASS_http_getwebinfo)
// 		if string(bs) == "ok" {
// 			return 0
// 		}
// 		portVO := new(NodeWebinfoVO)
// 		err := json.Unmarshal(bs, portVO)
// 		if err != nil {
// 			return 0
// 		}
// 		return portVO.WebPort
// 	}
// 	return 0
// }
