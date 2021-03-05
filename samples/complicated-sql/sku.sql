-- +gen:sql v=0.3

-- 根据商品id, 获取这个商品的所有sku id和基本信息
-- +gen:method Gets
SELECT
	sku.id,
	ifnull(se.ref_id, '') RefId,
	sku.status,
	ifnull(se.origin_price, 0) Price,
	sku.sku_uid Uid
FROM
	sku
LEFT JOIN
	 sku_ezbuy se ON se.sku_id=sku.id
JOIN
	sku_factor sf ON sf.sku_id=sku.id
WHERE
	sku.product_id=${id} AND (sku.status=1 OR sku.status=2)
    AND sf.sku_selling_method=2;
-- +gen:end


-- 根据商品id, 获取其所有的sku属性和属性值
-- 属性排序和属性值排序会单独拿出来, 在程序中进行排序
-- sku_id会拿出来, 用于将属性关联到上面通过Gets获取的sku
-- +gen:method GetProps
SELECT
	sai.sku_id,
	sai.attributes_id AttrId,
	sai.attributes_values_id ValId,
	attr.attributes_sort_order AttrSort,
	val.attributes_values_sort_order ValSort
FROM
	sku_attributes_info sai
JOIN
	`attributes` attr ON attr.id=sai.attributes_id
JOIN
	attributes_values val ON val.id=sai.attributes_values_id
WHERE
	products_id=${id} AND attr.status=1 AND val.status=1;
-- +gen:end



-- 获取属性关联产品图的关系. 对于新商品中心, 属性图实际上就是
-- 产品图, 不过通过这张关联表创建关联关系. 这里获取这个关系以
-- 确定属性图.
-- +gen:method GetPropImgs
SELECT
	attr_value_ids_json Pair,
	images_id ImgId
FROM
	products_attributes_images
WHERE
	products_id=${id} AND status=1;
-- +gen:end


-- 获取sku属性的多语言标题
-- +gen:method GetPropTitles dyn=true
SELECT
	ad.attributes_id Id,
	ad.languages_id LangId,
	ad.attributes_title Title
FROM
	attributes_desc ad
WHERE
	ad.attributes_id IN (
	  %{for id in attrIds join ','}${id}%{endfor}
	)
	AND languages_id IN (#{langCodes});
-- +gen:end

-- 获取sku属性值的多语言标题
-- +gen:method GetPropValTitles dyn=true
SELECT
	avd.attributes_values_id Id,
	avd.languages_id LangId,
	avd.attributes_values_title Title
FROM
	attributes_values_desc avd
WHERE
	avd.attributes_values_id IN (
	  %{for id in valIds join ','}${id}%{endfor}
	)
	AND languages_id IN (#{langCodes});
-- +gen:end


-- 获取sku的销售模式(经销/库存)
-- +gen:method SellType
SELECT
	sku_selling_method
FROM
	sku_factor
WHERE
	sku_id=${skuId}
-- +gen:end

