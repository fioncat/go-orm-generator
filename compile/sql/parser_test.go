package sql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fioncat/go-gendb/compile/base"
)

func doParseSql(sqls [][]string) {
	fmt.Println("==========================")
	for _, lines := range sqls {
		tagLine := lines[0]
		tag, err := base.ParseTag(commPrefix, tagLine)
		if err != nil {
			fmt.Printf("parse tag failed: %v\n", err)
			return
		}
		p, err := newSqlParser(tag)
		if err != nil {
			fmt.Printf("newSqlParser failed: %v\n", err)
			return
		}

		lines = lines[1:]
		for idx, line := range lines {
			p.Next(idx, line, nil)
		}

		v := p.Get()
		switch v.(type) {
		case error:
			fmt.Println(v.(error).Error())
			return

		case *Method:
			m := v.(*Method)
			fmt.Printf("Method name=%s, inter=%s\n",
				m.Name, m.Inter)
			if m.Dyn {
				for idx, dp := range m.Dps {
					phs := make([]string, len(dp.State.phs))
					for idx, ph := range dp.State.phs {
						phs[idx] = ph.String()
					}
					fmt.Printf("dp %d: sql=%s, phs=[%s]\n",
						idx, dp.State.Sql, strings.Join(phs, ","))
					switch dp.Type {
					case DynamicTypeConst:

					case DynamicTypeIf:
						fmt.Printf("\tIfCond=%s\n", dp.IfCond)

					case DynamicTypeFor:
						fmt.Printf("\tslice=%s, ele=%s, join=%s\n",
							dp.ForSlice, dp.ForEle, dp.ForJoin)
					}
				}
			} else {
				phs := make([]string, len(m.State.phs))
				for idx, ph := range m.State.phs {
					phs[idx] = ph.String()
				}
				fmt.Printf("no-dyn: sql=%s, phs=[%s]\n",
					m.State.Sql, strings.Join(phs, ","))
			}
			fmt.Printf("IsExec = %v\n", m.Exec)
			if !m.Exec {
				for _, f := range m.Fields {
					fmt.Printf("\tfname=%s, table=%s, alias=%s\n",
						f.Name, f.Table, f.Alias)
				}
			}
			fmt.Println("==========================")
		}
	}
}

func TestSql0(t *testing.T) {
	sqls := [][]string{
		{
			"-- +gen:method FindById dyn=false",
			"SELECT id, name, age",
			"FROM user",
			"WHERE id=${id}",
		},
		{
			"-- +gen:method userOper.Add",
			"INSERT INTO `user`(`id`, `name`, `age`)",
			"VALUES (${u.Id}, ${u.Name}, ${u.Age})",
		},
	}
	doParseSql(sqls)
}

func TestSql1(t *testing.T) {
	sqls := [][]string{
		{
			"-- +gen:method Get",
			"SELECT",
			"  pp.status `PlatformStatus`,",
			"  pp.id PlatformId,",
			"  p.status BaseStatus,",
			"  p.products_uid Uid,",
			"  ifnull(bp.origin_name, '') BrandName,",
			"  ifnull(pe.ezbuy_category_id, 0) Pcid,",
			"  ifnull(pe.ref_id, '') RefId,",
			"  ifnull(pe.origin_code, '') Region,",
			"  ifnull(mme.manufacturers_id, 0) ManufactureId,",
			"  ifnull(mme.manufacturers_name, '') Manufacture,",
			"  ifnull(mme.vendor_id_ezbuy, 0) VendorId,",
			"  ifnull(mme.vendor_ezbuy, '') VendorName,",
			"  ifnull(pe.text_desc, '') Text,",
			"  ifnull(pe.delivery_method, 0) DeliveryMethod,",
			"  ifnull(pe.location, '') Location,",
			"  ifnull(pe.is_sensitive_item, 0) IsSens,",
			"  ifnull(pdi.sell_types_id, 0) SellType",
			"",
			"FROM products p",
			"LEFT JOIN products_to_brands ptb ON ptb.products_id=p.id AND ptb.status=1",
			"LEFT JOIN brands_pub bp ON bp.brand_id=ptb.brands_id AND bp.status=1",
			"LEFT JOIN products_ezbuy pe ON p.id=pe.products_id",
			"LEFT JOIN manufactures_mapping_ezbuy mme ON p.manufacturers_id=mme.manufacturers_id",
			"LEFT JOIN products_platform pp ON pp.products_id=p.id",
			"LEFT JOIN products_detail_info pdi ON pdi.products_id=p.id",
			"WHERE p.id=${id};",
		},
		{
			"-- +gen:method GetImgs",
			"SELECT",
			"  pi.id,",
			"  pip.images_url Url,",
			"  pi.is_default,",
			"  pi.sort_order",
			"FROM products_images_platform pip",
			"JOIN products_images pi ON pip.products_images_id=pi.id AND pi.status=1",
			"JOIN products_images_tags pit ON pit.products_images_id=pip.products_images_id AND pit.status=1 AND pit.tag_id=20",
			"WHERE pip.products_platform_id=${pid};",
		},
		{
			"-- +gen:method GetTitle",
			"SELECT",
			"  lang.code Code,",
			"  ifnull(pc.products_name,'') Name",
			"FROM products_cms pc",
			"JOIN languages lang ON lang.languages_id=pc.languages_id",
			"WHERE pc.products_id=${id} AND lang.code IN ('en','cn','tw','th');",
		},
		{
			"-- +gen:method GetDescImgs",
			"SELECT",
			"  pdp.cdn_url Url,",
			"  pd.sort_order",
			"  FROM products_descriptions_platform pdp",
			"JOIN products_descriptions pd ON pdp.products_descriptions_id=pd.id",
			"WHERE pd.`type`='img' AND pd.status=1 AND pd.products_id=${id};",
		},
	}
	doParseSql(sqls)
}

func TestDyn0(t *testing.T) {
	sqls := [][]string{
		{
			"-- +gen:method AddBatch dyn=true",
			"INSERT INTO `user`(`name`, `age`, `phone`, `email`)",
			"VALUES",
			"%{for u in users join ','}",
			"(${u.Name}, ${u.Age}, ${u.Phone}, ${u.Email})",
			"%{endfor}",
		},
		{
			"-- +gen:method GetPropTitles dyn=true",
			"SELECT",
			" ad.attributes_id Id,",
			" ad.languages_id LangId,",
			" ad.attributes_title Title",
			"FROM attributes_desc ad",
			"WHERE ad.attributes_id IN (%{for id in attrIds join ','}${id}%{endfor})",
			"      AND languages_id IN (#{langCodes});",
		},
	}
	doParseSql(sqls)
}
