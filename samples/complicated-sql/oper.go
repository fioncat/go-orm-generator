// +gen:sql v=0.3 conn=eztest
// +gen:sql db_use=getDB()

package db

import "database/sql"

func getDB() *sql.DB {
	return nil
}

// +gen:BaseOper base.sql
type baseOper interface {
	// +gen:auto-ret
	Get(id int64) (*Base, error)

	// +gen:auto-ret
	GetImgs(pid int64) ([]*Img, error)

	// +gen:auto-ret
	GetTitle(id int64) ([]*Title, error)

	// +gen:auto-ret
	GetDescAttrs(id int64) ([]*DescAttr, error)

	// +gen:auto-ret
	GetDescImgs(id int64) ([]*DescImg, error)

	// +gen:auto-ret
	GetDescVideos(id int64) ([]*DescVideo, error)

	EzGpid(id int64) (int64, error)

	IsSenstive(id int64) (bool, error)

	ManufactureId(id int64) (int32, error)

	SellerType(id int64) (int32, error)

	Platform(id int64) (string, error)

	// +gen:auto-ret
	Warehouse(id int64) ([]WarehouseFee, error)

	// +gen:auto-ret
	ShipmentTypes(uid string) ([]ShipmentType, error)
}

// +gen:FestivalOper festival.sql
type festivalOper interface {
	// +gen:auto-ret
	List(now string) ([]Festival, error)

	Get(now string) ([]Festival, error)
}

// +gen:LangOper lang.sql
type langOper interface {
	// +gen:auto-ret
	GetLangs() ([]*Lang, error)
}

// +gen:SkuOper sku.sql
type skuOper interface {
	// +gen:auto-ret
	Gets(id int64) ([]*Sku, error)

	// +gen:auto-ret
	GetProps(id int64) ([]*SkuProp, error)

	// +gen:auto-ret
	GetPropImgs(id int64) ([]*PropImg, error)

	// +gen:auto-ret
	GetPropTitles(attrIds []int64, langCodes string) ([]*PropTitle, error)

	GetPropValTitles(valIds []int64, langCodes string) ([]*PropTitle, error)

	SellType(skuId int64) (int32, error)
}
