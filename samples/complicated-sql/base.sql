-- +gen:sql v=0.3

-- Get Base Product Info
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
FROM products p
LEFT JOIN products_to_brands ptb ON ptb.products_id=p.id AND ptb.status=1
LEFT JOIN brands_pub bp ON bp.brand_id=ptb.brands_id AND bp.status=1
LEFT JOIN products_ezbuy pe ON p.id=pe.products_id
LEFT JOIN manufactures_mapping_ezbuy mme ON p.manufacturers_id=mme.manufacturers_id
LEFT JOIN products_platform pp ON pp.products_id=p.id
LEFT JOIN products_detail_info pdi ON pdi.products_id=p.id
WHERE p.id=${id};
-- +gen:end



-- +gen:method GetImgs
SELECT
  pi.id,
  pip.images_url Url,
  pi.is_default,
  pi.sort_order
FROM products_images_platform pip
JOIN products_images pi ON pip.products_images_id=pi.id AND pi.status=1
JOIN products_images_tags pit ON pit.products_images_id=pip.products_images_id AND pit.status=1 AND pit.tag_id=20
WHERE pip.products_platform_id=${pid};
-- +gen:end


-- +gen:method GetTitle
SELECT
  lang.code Code,
  ifnull(pc.products_name,'') Name
FROM products_cms pc
JOIN languages lang ON lang.languages_id=pc.languages_id
WHERE pc.products_id=${id} AND lang.code IN ('en','cn','tw','th');
-- +gen:end


-- +gen:method GetDescAttrs
SELECT
  ifnull(attr.name,'') Name,
  val.value Value,
  pda.sort_order SortOrder
FROM products_desc_attributes pda
JOIN desc_attributes attr ON attr.id=pda.desc_attributes_id
JOIN desc_attributes_values val ON val.id=pda.desc_attributes_values_id
WHERE pda.products_id=${id} AND attr.status=1 AND val.status=1 AND pda.status=1;
-- +gen:end



-- +gen:method GetDescImgs
SELECT
  pdp.cdn_url Url,
  pd.sort_order
FROM products_descriptions_platform pdp
JOIN products_descriptions pd ON pdp.products_descriptions_id=pd.id
WHERE pd.`type`='img' AND pd.status=1 AND pd.produc
-- +gen:end


-- +gen:method GetDescVideos
SELECT
  pvp.video_cdn_url Url,
  ifnull(pvp.thumbnail_cdn_url, '') Preview,
  pv.media_type,
  pv.show_flag
FROM products_videos_platform pvp
JOIN products_videos pv ON pvp.products_videos_id=pv.id
WHERE pv.status=1 AND pv.products_id=${id};
-- +gen:end


-- +gen:method EzGpid
SELECT
  ppe.old_products_id
FROM products_platform pp
JOIN products_platform_extend ppe ON pp.id=ppe.products_platform_id
WHERE pp.products_id=${id};
-- +gen:end


