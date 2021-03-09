-- +gen:sql v=0.3


-- 获取商品的基础信息
-- +gen:method Get
SELECT
	pp.status PlatformStatus,
	pp.id PlatformId,
	p.status BaseStatus,
	p.products_uid Uid,
	ifnull(bp.origin_name, '') BrandName,
	ifnull(pe.ezbuy_category_id,0) Pcid,
	ifnull(pe.ref_id, '') RefId,
	ifnull(pe.origin_code, '') Region,
	ifnull(mme.manufacturers_id, 0) ManufactureId,
	ifnull(mme.manufacturers_name, '') Manufacture,
	ifnull(mme.vendor_id_ezbuy, 0) VendorId,
	ifnull(mme.vendor_ezbuy, '') VendorName,
	ifnull(pe.text_desc, '') Text,
	ifnull(pe.delivery_method, 0) DeliveryMethod,
	ifnull(pe.location, '') Location,
	ifnull(pe.is_sensitive_item, 0) IsSens,
	ifnull(pdi.sell_types_id, 0) SellType
FROM
	products p
LEFT JOIN
	 products_to_brands ptb
	 ON ptb.products_id=p.id AND ptb.status=1
LEFT JOIN
	 brands_pub bp
	 ON bp.brand_id=ptb.brands_id AND bp.status=1
LEFT JOIN
	 products_ezbuy pe
	 ON p.id=pe.products_id
LEFT JOIN
	 manufactures_mapping_ezbuy mme
	 ON p.manufacturers_id=mme.manufacturers_id
LEFT JOIN
	 products_platform pp
	 ON pp.products_id=p.id
LEFT JOIN
	 products_detail_info pdi
	 ON pdi.products_id=p.id
WHERE
	p.id=${id};
-- +gen:end


-- 获取产品图, 其中包括了属性图, 属性图的关联关系通过
-- sku中的GetPropImgs获取. 因此这里需要把id给拿到
-- 另外, sort_order被拿出来, 在程序中执行排序以减轻数据库压力.
-- +gen:method GetImgs
SELECT
	pi.id,
	pip.images_url Url,
	pi.is_default,
	pi.sort_order
FROM
	products_images_platform pip
JOIN
	products_images pi
	ON pip.products_images_id=pi.id AND pi.status=1
JOIN
	products_images_tags pit
	ON pit.products_images_id=pip.products_images_id
	AND pit.status=1 AND pit.tag_id=20
WHERE
	pip.products_platform_id=${pid};
-- +gen:end


-- 获取某个产品的多语言翻译的标题
-- +gen:method GetTitle
SELECT
	lang.code Code,
	ifnull(pc.products_name,'') Name
FROM
	products_cms pc
JOIN
	languages lang
	ON lang.languages_id=pc.languages_id
WHERE
	pc.products_id=${id}
	AND lang.code IN ('en','cn','tw','th');
-- +gen:end


-- 获取产品的描述属性
-- +gen:method GetDescAttrs
SELECT
	ifnull(attr.name,'') Name,
	val.value Value,
	pda.sort_order SortOrder
FROM
	products_desc_attributes pda
JOIN
	desc_attributes attr
	ON attr.id=pda.desc_attributes_id
JOIN
	desc_attributes_values val
	ON val.id=pda.desc_attributes_values_id
WHERE
	pda.products_id=${id} AND attr.status=1
	AND val.status=1 AND pda.status=1;

-- +gen:end


-- 获取产品的描述图
-- +gen:method GetDescImgs
SELECT
	pdp.cdn_url Url,
	pd.sort_order
FROM
	products_descriptions_platform pdp
JOIN
	products_descriptions pd
	ON pdp.products_descriptions_id=pd.id
WHERE
	pd.`type`='img' AND pd.status=1
	AND pd.products_id=${id};

-- +gen:end


-- 获取产品的描述视频
-- +gen:method GetDescVideos
SELECT
	pvp.video_cdn_url Url,
	ifnull(pvp.thumbnail_cdn_url, '') Preview,
	pv.media_type,
	pv.show_flag
FROM
	products_videos_platform pvp
JOIN
	products_videos pv
	ON pvp.products_videos_id=pv.id
WHERE
	pv.status=1 AND pv.products_id=${id};

-- +gen:end


-- 根据新商品中心id关联获取ezbuy的gpid
-- +gen:method EzGpid
SELECT
	ppe.old_products_id
FROM
	products_platform pp
JOIN
	products_platform_extend ppe
	ON pp.id=ppe.products_platform_id
WHERE
	pp.products_id=${id};

-- +gen:end


-- 获取商品是否是敏感品
-- +gen:method IsSenstive
SELECT
	ifnull(is_sensitive_item, 0)
FROM
	products_ezbuy
WHERE
	products_id=${id};

-- +gen:end


-- 获取某个商品的供应商id
-- +gen:method ManufactureId
SELECT
	ifnull(manufacturers_id, 0)
FROM
	products
WHERE
	id=${id};

-- +gen:end


-- 获取某个商品的销售模式(经销/库存)
-- +gen:method SellerType
SELECT
	sell_types_id
FROM
	products_detail_info
WHERE
	products_id=${id}

-- +gen:end


-- 获取某个商品的平台类型(ezbuy/litb)
-- +gen:method Platform
SELECT
	platform
FROM
	products_platform
WHERE
	products_id=${id};

-- +gen:end


-- 获取某个商品的仓库和运费
-- +gen:method Warehouse
SELECT
	warehouse Name,
	fee
FROM
	products_ezbuy_freight
WHERE
	products_id=${id}

-- +gen:end


-- 获取某个商品的敏感运输方式
-- +gen:method ShipmentTypes
SELECT
	pst.sensitive_country Catalog,
	pst.sensitive_type Type
FROM
	products_shipping_type pst
WHERE
	pst.products_uid=${uid}

-- +gen:end
