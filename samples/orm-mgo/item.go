// +gen:orm-mgo v=0.3

package item

// +gen:orm db=true name=Base
type _base struct {
	// +gen:orm flags=[unique]
	Gpid int64

	ParentGpid int64

	InSale bool

	Attrs []*_skuAttr
}

// +gen:orm name=SkuAttr
type _skuAttr struct {
	PropId        string
	PropValueId   string
	Img           string
	PropName      map[string]string
	PropValueName map[string]string
}
