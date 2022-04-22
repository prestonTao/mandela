package config

var (
	VNODE_get_neighbor_vnode_tiker = []int64{60 * 10} //定时获取邻居节点的虚拟节点地址
	VNODE_tiker_sync_logical_vnode = []int64{60 * 10} //每个虚拟节点定时从自己的逻辑节点查询逻辑节点
)
