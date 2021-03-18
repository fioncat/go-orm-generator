// +gen:common v=0.3

package item

import "gopkg.in/mgo.v2/bson"

// +gen:deepcopy
type Base struct {
	// +gen:deepcopy set='$0'
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Gpid     int64         `bson:"Gpid" json:"Gpid"`
	Region   int32         `bson:"Region" json:"Region"`
	Vendor   string        `bson:"Vendor" json:"Vendor"`
	VendorId string        `bson:"VendorId" json:"VendorId"`
	InSale   bool          `bson:"InSale" json:"InSale"`
	Pcid     int32         `bson:"Pcid" json:"Pcid"`
	Brand    string        `bson:"Brand" json:"Brand"`

	// +gen:deepcopy set=$0
	Catalog        map[string]BaseCatalog `bson:"Catalog" json:"Catalog"`
	Skus           []int64                `bson:"Skus" json:"Skus"`
	RemovedSkus    []int64                `bson:"RemovedSkus" json:"RemovedSkus"`
	Platform       int32                  `bson:"Platform" json:"Platform"`
	RefId          string                 `bson:"RefId" json:"RefId"`
	EstWeight      int32                  `bson:"EstWeight" json:"EstWeight"`
	EstSetWeight   int32                  `bson:"EstSetWeight" json:"EstSetWeight"`
	EstPrimeWeight int32                  `bson:"EstPrimeWeight" json:"EstPrimeWeight"`
	MinWeight      int32                  `bson:"MinWeight" json:"MinWeight"`
	AllDcids       []int32                `bson:"AllDcids" json:"AllDcids"`
	DcidPath       [][]int32              `bson:"DcidPath" json:"DcidPath"`
	DcidPathMap    map[string][][]int32   `bson:"DcidPathMap" json:"DcidPathMap"`
	PlatformCid    []int64                `bson:"PlatformCid" json:"PlatformCid"`
	PlatformScid   []string               `bson:"PlatformScid" json:"PlatformScid"`
	PlatformSite   string                 `bson:"PlatformSite" json:"PlatformSite"`
	PlatformUrl    string                 `bson:"PlatformUrl" json:"PlatformUrl"`
	PlatformShop   string                 `bson:"PlatformShop" json:"PlatformShop"`
	InternalUrl    string                 `bson:"InternalUrl" json:"InternalUrl"`

	// +gen:deepcopy set=$0
	Acts               []ActivityItem `bson:"Acts" json:"Acts"`
	CreateDate         int64          `bson:"CreateDate" json:"CreateDate"`
	UpdateDate         int64          `bson:"UpdateDate" json:"UpdateDate"`
	ListingScore       int32          `bson:"ListingScore" json:"ListingScore"`
	Pvids              []int32        `bson:"Pvids" json:"Pvids"`
	IsPremium          bool           `bson:"IsPremium" json:"IsPremium"`
	ForcePremium       int32          `bson:"ForcePremium" json:"ForcePremium"`
	IsValuable         bool           `bson:"IsValuable" json:"IsValuable"`
	IsFragile          bool           `bson:"IsFragile" json:"IsFragile"`
	IsSplittable       bool           `bson:"IsSplittable" json:"IsSplittable"`
	IsThirdPartySeller bool           `bson:"IsThirdPartySeller" json:"IsThirdPartySeller"`
	Version            int64          `bson:"Version" json:"Version"`

	// +gen:deepcopy set='$0'
	AllBanRules    []bson.ObjectId `bson:"AllBanRules" json:"AllBanRules"`
	NoEzbuy        bool            `bson:"NoEzbuy" json:"NoEzbuy"`
	From           []int32         `bson:"From" json:"From"`
	InSaleTimeFrom int64           `bson:"InSaleTimeFrom" json:"InSaleTimeFrom"`
	InSaleTimeTo   int64           `bson:"InSaleTimeTo" json:"InSaleTimeTo"`
	MinPrice       int64           `bson:"MinPrice" json:"MinPrice"`
	MaxPrice       int64           `bson:"MaxPrice" json:"MaxPrice"`
	DeliveryMethod int32           `bson:"DeliveryMethod" json:"DeliveryMethod"`
	SaleType       int32           `bson:"SaleType" json:"SaleType"`
	isNew          bool
}
