package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

/**
命令行: jsontogo -n myStruct -s '{"a":[{"b":"bb"}],"c":"d"}'
*/

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
		dest += fmt.Sprintf(`%s:"%s" `, tag, key)
	}
	dest = dest[0:len(dest)-1] + "`"
	return dest
}

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

/* ProduceMapTag 转换
obj: 数据
tag: 标签
level: 第几级的数据
needReturn: 是否需要回车换行
*/
func ProduceMapTag(obj map[string]interface{}, tag string, level int, needReturn bool) string {
	newDefineCode := "{\n"
	firstTabs := ""
	endTabs := ""

	for i := 0; i < level; i++ {
		firstTabs += "\t"
		if i < level-1 {
			endTabs += "\t"
		}
	}
	for k, v := range obj {
		// fmt.Println(k, v)
		s := reflect.ValueOf(v)
		switch v.(type) {

		case []string:
			newDefineCode += fmt.Sprintf("\t%s string %s\n", FirstUpper(k), multiTags(tag, k)) // example value: %+v
			break

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
	return fmt.Sprintf("%s%s}", newDefineCode, endTabs)
}

/* ProduceMapTagEach 转换
obj: 数据
tag: 标签
needReturn: 是否需要回车换行
*/
func ProduceMapTagEach(obj map[string]interface{}, tag string) string {
	newDefineCode := "{\n"
	subStructs := make([]string, 0)

	for k, v := range obj {
		// fmt.Println(k, v)
		s := reflect.ValueOf(v)
		switch v.(type) {
		case map[string]interface{}:
			// fmt.Println("map[string]interface{}", k, v)
			newDefineCode += fmt.Sprintf("%s\t%s %s\n", FirstUpper(k), FirstUpper(k), multiTags(tag, k))                           // example value: %+v
			subDefineCode := fmt.Sprintf("type %s struct %s\n", FirstUpper(k), ProduceMapTagEach(v.(map[string]interface{}), tag)) // example value: %+v
			subStructs = append(subStructs, subDefineCode)
			break
		case []map[string]interface{}:
			// fmt.Println("[]map[string]interface{}", k, v)
			newDefineCode += fmt.Sprintf("%s\t[]%s %s\n", FirstUpper(k), FirstUpper(k), multiTags(tag, k)) // example value: %+v
			subDefineCode := fmt.Sprintf("type %s struct %s\n", FirstUpper(k), ProduceMapTagEach(v.([]map[string]interface{})[0], tag))
			subStructs = append(subStructs, subDefineCode)
			break
		case []interface{}:
			// fmt.Println("interface{}", k, v)
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
			newDefineCode += fmt.Sprintf("%s\t[]struct %s %s\n", FirstUpper(k), FirstUpper(k), multiTags(tag, k)) // example value: %+v
			subDefineCode := fmt.Sprintf("type %s struct %s\n", FirstUpper(k), ProduceMapTagEach(otherSrc, tag))
			subStructs = append(subStructs, subDefineCode)
			break
		default:
			// fmt.Println("default", v)
			newDefineCode += fmt.Sprintf("%s\t%s\t%s\n", FirstUpper(k), s.Type(), multiTags(tag, k)) // example value: %+v
			break
		}
	}

	for _, s := range subStructs {
		fmt.Println(s)
	}
	return fmt.Sprintf("%s}", newDefineCode)
}

func main() {

	src := flag.String("s", `{"a":"b","c":{"e":"f","g":{"i":"j"}}}`, "The json string ")
	name := flag.String("n", "AutoGenerated", "structural name")
	tag := flag.String("t", "json", "tag name, many tags can split with '|'")
	individual := flag.Bool("i", false, "individual each struct")

	flag.Parse()

	s := []byte(*src)
	obj := make(map[string]interface{})
	err := json.Unmarshal(s, &obj)
	if err != nil {
		fmt.Println("json Unmarshal error", err, *src)
	} else {

		head := fmt.Sprintf(`type %s struct `, FirstUpper(*name))
		if *individual {
			fmt.Println(head + ProduceMapTagEach(obj, *tag))
		} else {
			fmt.Println(head + ProduceMapTag(obj, *tag, 1, true))
		}
	}
}
