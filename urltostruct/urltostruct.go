package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"unicode"
)

func FirstUpper(src string) string {
	first := true
	dest := ""
	for _, s := range src {
		if first {
			if !unicode.IsUpper(s) {
				dest += string(strings.ToUpper(string(s)))
			} else {
				dest += string(s)
			}
			first = false
		} else {
			if string(s) == "_" {
				first = true
			} else {
				dest += string(s)
			}
		}
	}

	return dest
}

func getTagName(currName, tag string) (newName string) {
	first := true
	for _, r := range currName {
		if unicode.IsUpper(r) {
			if first {
				newName = fmt.Sprintf("%s%s", newName, strings.ToLower(string(r)))
				first = false
			} else {
				newName = fmt.Sprintf("%s_%s", newName, strings.ToLower(string(r)))
			}
		} else {
			newName = fmt.Sprintf("%s%s", newName, string(r))
		}
	}
	newName = fmt.Sprintf("`%s:\"%s\"`", tag, newName)
	return
}

func ProduceStructTag(obj interface{}, tag string) string {
	var newDefineCode string
	s := reflect.ValueOf(obj)
	if reflect.TypeOf(obj).Kind() == reflect.Map {
		return ProduceMapTag(obj, tag, 1, false)
	}
	newDefineCode = fmt.Sprintf("type %s struct {\n", s.Type().String())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		n := s.Type().Field(i).Name
		newDefineCode = fmt.Sprintf("%s\t%s\t%s\t\t%s\n",
			newDefineCode,
			n,
			f.Type(),
			getTagName(n, tag))
	}
	newDefineCode = fmt.Sprintf("%s}\n", newDefineCode)
	return newDefineCode
}

func multiTags(tags, key string) string {
	tagss := strings.Split(tags, "|")
	dest := "`"
	for _, tag := range tagss {
		if tag == "valid" {
			dest += fmt.Sprintf(`%s:"required" `, tag)
		} else {
			dest += fmt.Sprintf(`%s:"%s" `, tag, key)
		}
	}
	dest = dest[0:len(dest)-1] + "`"
	return dest
}

/* ProduceMapTag 转换
obj: 数据
tag: 标签
level: 第几级的数据
needReturn: 是否需要回车换行
*/
func ProduceMapTag(obj interface{}, tag string, level int, needReturn bool) string {
	// fmt.Println(reflect.TypeOf(obj))
	newDefineCode := "{\n"
	switch obj.(type) {
	case url.Values:
		if u, ok := obj.(url.Values); ok {
			// fmt.Println(u, ok)
			for k := range u {
				newDefineCode += fmt.Sprintf("\t%s string %s\n", FirstUpper(k), multiTags(tag, k)) // example value: %+v
			}
		}
		break

	case map[string]interface{}:
		firstTabs := ""
		endTabs := ""

		for i := 0; i < level; i++ {
			firstTabs += "\t"
			if i < level-1 {
				endTabs += "\t"
			}
		}
		for k, v := range obj.(map[string]interface{}) {
			// fmt.Println(k, v)
			s := reflect.ValueOf(v)
			switch v.(type) {
			case map[string]interface{}:
				newDefineCode = fmt.Sprintf("%s%s%s\tstruct %s %s\n", newDefineCode, firstTabs, FirstUpper(k), ProduceMapTag(v.(map[string]interface{}), tag, level+1, false), multiTags(tag, k)) // example value: %+v
				break
			case []map[string]interface{}:
				newDefineCode = fmt.Sprintf("%s%s%s\t[]struct %s %s\n", newDefineCode, firstTabs, FirstUpper(k), ProduceMapTag(v.([]map[string]interface{})[0], tag, level+1, false), multiTags(tag, k)) // example value: %+v
				break
			case []interface{}:
				b, e := json.Marshal(v.([]interface{})[0])
				if e != nil {
					fmt.Println(e)
					break
				}
				otherSrc := make(map[string]interface{}, 0)
				e = json.Unmarshal(b, &otherSrc)
				if e != nil {
					fmt.Println(e)
					break
				}
				newDefineCode = fmt.Sprintf("%s%s%s\t[]struct %s %s\n", newDefineCode, firstTabs, FirstUpper(k), ProduceMapTag(otherSrc, tag, level+1, false), multiTags(tag, k)) // example value: %+v
				break
			default:
				newDefineCode = fmt.Sprintf("%s%s%s\t%s\t%s\n", newDefineCode, firstTabs, FirstUpper(k), s.Type(), multiTags(tag, k)) // example value: %+v
				break
			}
		}
		if needReturn {
			return fmt.Sprintf("%s%s}\n", newDefineCode, endTabs)
		}
		break

	default:
		fmt.Println(reflect.TypeOf(obj))
		break
	}
	return fmt.Sprintf("%s}", newDefineCode)
}

func main() {

	src := flag.String("s", `http://testinottpay.api.mgtv.com/v1/epg5/getVodPlayUrl?android_sdk_ver=19&license=ZgOOgo5MjkyOTDsGS3xLDbSqIIe0Sw6Vqg0gSzt7mXZLSw0gtA2qe3aZqjs7O4e0BaoFjkyOTI5MZgOOgg==&net_id=&quality=1&mac_id=40-EA-CE-03-76-CE&part_id=5524879&mod=p201_iptv&pre=0&uuid=2c14080d660d4eadbb96fe2618b6a4c4&abt=&version=5.9.002.200.3.KS_TVAPP.0.0_Release&business_id=2000004&mf=amlogic&svrip=&os_ver=4.4.2&model_code=p201_iptv&platform=3&rom_version=&ticket=5NAX9106XBBUTCIPFBE9&force_avc=0&app_type=1&device_id=c028294365421a3962cdbe2296493deb3ccc5473&dcp_id=0&buss_id=1000014&cur_play_id=100024904&_support=00100101011&time_zone=GMT+08:00&channel_code=KS`, "The url string ")
	name := flag.String("n", "AutoGenerated", "structural name")
	tag := flag.String("t", "xml", "tag name, many tags can split with '|'")
	// individual := flag.Bool("i", false, "individual each struct")

	flag.Parse()
	u, err := url.Parse(*src)
	if err != nil {
		fmt.Println(err)
		return
	}
	v, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(v)
	// for kk, vv := range v {
	// fmt.Println(kk, reflect.ValueOf(vv[0]).Type())
	// }
	fmt.Println(fmt.Sprintf("type %s struct %s\n", *name, ProduceStructTag(v, *tag)))
}