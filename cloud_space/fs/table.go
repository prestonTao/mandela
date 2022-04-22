package fs

type Key0 struct {
	Key
}
type Key1 struct {
	Key
}
type Key2 struct {
	Key
}
type Key3 struct {
	Key
}
type Key4 struct {
	Key
}
type Key5 struct {
	Key
}
type Key6 struct {
	Key
}
type Key7 struct {
	Key
}
type Key8 struct {
	Key
}
type Key9 struct {
	Key
}
type Key10 struct {
	Key
}

//type KeyInterface interface {
//	Set(key string, value []byte)
//}

type Key struct {
	Id     uint64 `xorm:"pk autoincr unique 'id'"` //id
	Key    string `xorm:"TEXT 'key'"`              //
	Status int    `xorm:"int 'status'"`            //块保存状态。1=未使用；2=已使用。
	Value  []byte `xorm:"Blob 'value'"`            //
}
