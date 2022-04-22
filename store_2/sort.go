/*
对在线共享者排序，更新时间越近，则优先提供下载，用于过滤下线的共享者
*/
package store

import (
	"sort"
)

type SU []*ShareUser

//Len()
func (su SU) Len() int {
	return len(su)
}

//Less():更新时间由高到底排序
func (su SU) Less(i, j int) bool {
	return su[i].UpdateTime > su[j].UpdateTime
}

//Swap()
func (su SU) Swap(i, j int) {
	su[i], su[j] = su[j], su[i]
}
func SortSU(su []*ShareUser) []*ShareUser {
	user := SU(su)
	sort.Sort(user)
	return user
}
