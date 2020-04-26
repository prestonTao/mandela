package core

import (
	"mandela/config"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"encoding/hex"
)

func GetNodeMachineID(recvid *nodeStore.AddressNet) int64 {
	message, ok, isSendSelf := message_center.SendP2pMsg(config.MSGID_search_node, recvid, nil)
	if isSendSelf || !ok {
		return 0
	}

	bs := flood.WaitRequest(config.CLASS_get_MachineID, hex.EncodeToString(message.Body.Hash), 0)
	if bs == nil {
		return 0
	}
	return utils.BytesToInt64(*bs)
}
