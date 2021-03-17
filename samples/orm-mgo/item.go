// +gen:orm-mgo v=0.3

package item

// +gen:orm db=true name=Base
type _base struct {
	// +gen:orm flags=[unique]
	Gpid int64

	// +gen:orm flags=[index]
	ParentGpid int64

	InSale bool

	Attrs []*_skuAttr

	// +gen:orm flags=[index]
	UpdateDate int64
}

// +gen:orm name=SkuAttr
type _skuAttr struct {
	PropId        string
	PropValueId   string
	Img           string
	PropName      map[string]string
	PropValueName map[string]string
}
