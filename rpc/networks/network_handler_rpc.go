package networks

import (
	"mandela/core/nodeStore"
	"mandela/rpc/model"
	"net/http"
)

/*
	获取本节点网络信息，包括节点地址，本节点是否是超级节点
*/
func NetworkInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	netAddr := nodeStore.NodeSelf.IdInfo.Id.B58String()
	isSuper := nodeStore.NodeSelf.IsSuper
	nodes := nodeStore.GetLogicNodes()
	ids := make([]string, 0)
	for _, one := range nodes {
		ids = append(ids, one.B58String())
	}

	otherNodes := nodeStore.GetOtherNodes()
	oids := make([]string, 0)
	for _, one := range otherNodes {
		oids = append(oids, one.B58String())
	}

	m := make(map[string]interface{})
	m["netaddr"] = netAddr
	m["issuper"] = isSuper
	m["superNodes"] = ids
	m["otherNodes"] = oids

	res, err = model.Tojson(m)
	return
}
