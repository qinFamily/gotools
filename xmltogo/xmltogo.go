package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	// github.com/beevik/etree
)

/**
命令行: ./xmltogo -n myStruct -s '<xml>
	<return_code><![CDATA[FAIL]]></return_code>
	<return_msg><![CDATA[系统繁忙,请稍后再试.]]></return_msg>
	<result_code><![CDATA[FAIL]]></result_code>
	<err_code><![CDATA[268458547]]></err_code>
	<err_code_des><![CDATA[系统繁忙,请稍后再试.]]></err_code_des>
	<mch_billno><![CDATA[0010010404201411170000046542]]></mch_billno>
	<mch_id>10010404</mch_id>
	<wxappid><![CDATA[wx6fa7e3bab7e15415]]></wxappid>
	<re_openid><![CDATA[onqOjjmM1tad-3ROpncN-yUfa6uI]]></re_openid>
	<total_amount>1</total_amount>
	</xml>'
*/

func multiTags(tags, key string) string {
	tagss := strings.Split(tags, "|")
	dest := "`"
	for _, tag := range tagss {
		dest += fmt.Sprintf(`%s:"%s,CDATA" `, tag, key)
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

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
	GB2312  = Charset("GB2312")
)

func ConvertByte2String(byte []byte, charset Charset) string {

	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
		break
	case GB2312:
		var decodeBytes, _ = simplifiedchinese.GBK.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
		break
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}

	return str
}

type Attribute struct {
	Name  string
	Value string
}

type Token struct {
	Name       string
	Attributes []Attribute
	Tokens     []Token
}

type Result struct {
	Root Token
}

var result = Result{}
var structName = "AutoGenerate"
var xmlResponse = make(map[string]string, 0)

func parse_token(decoder *xml.Decoder, tt xml.Token, mytoken *Token, needKuohu bool) {
	// returnStr := "type " + structName + " struct {\n"

	token := tt.(xml.StartElement)
	name := token.Name.Local
	mytoken.Name = name

	isData := false
	// returnStr += fmt.Sprintf("\t%s xml.Name `xml:\"%s\"`\n", FirstUpper(name), name)

	for _, attr := range token.Attr {
		attrName := attr.Name.Local
		attrValue := attr.Value

		attr := Attribute{Name: attrName, Value: attrValue}

		mytoken.Attributes = append(mytoken.Attributes, attr)
	}

	var t xml.Token
	var err error

	for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {

		switch t.(type) {
		case xml.StartElement:
			// fmt.Println("xml.StartElement", "mytoken.Name", mytoken.Name)
			// 处理元素开始（标签）
			mytoken.Tokens = append(mytoken.Tokens, Token{})
			parse_token(decoder, t, &mytoken.Tokens[len(mytoken.Tokens)-1], true)
			break
		case xml.EndElement:
			// 处理元素结束（标签）
			// fmt.Println("xml.EndElement", "mytoken.Name", mytoken.Name)
			break
		case xml.CharData:
			// 处理元素结束（标签）
			b, ok := t.(xml.CharData)
			if !ok {
				fmt.Println("xml.CharData ok", ok)
				break
			}
			if bytes.Equal(b, []byte{10, 9}) {
				break
			}
			isData = true
			xmlResponse[mytoken.Name] = "xml.Data"
			// returnStr += fmt.Sprintf("\t%s string `xml:\"%s,CDATA\"`\n", FirstUpper(mytoken.Name), mytoken.Name)
			// fmt.Println("xml.CharData", "mytoken.Name", mytoken.Name, string(b))
		default:
			// fmt.Println(reflect.TypeOf(t).Kind())
			// fmt.Println("default")
			break
		}
		if needKuohu {
			// returnStr += "}\n"
		}
		// fmt.Println(returnStr)
	}
	if !isData {
		xmlResponse[mytoken.Name] = "xml.Root"
	}
}

func parse(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return parseIO(file)
}

func parseIO(ior io.Reader) bool {

	var t xml.Token

	decoder := xml.NewDecoder(ior)
	t, decerr := decoder.Token()
	if decerr != nil {
		fmt.Println(decerr)
		return false
	}
	parse_token(decoder, t, &result.Root, true)
	// fmt.Println("result", result.Root)
	return true
}

func genlist(n string) []string {
	num, _ := strconv.Atoi(n)
	ret := make([]string, num)
	for i := 0; i < num; i++ {
		ret[i] = strconv.Itoa(i)
	}
	return ret
}

func iconv(str string) string {
	return ConvertByte2String([]byte(str), "gb2312")
}

func output(src string, des string) bool {

	file, err := os.Create(des)
	if err != nil {
		fmt.Println(err)
		return false
	}

	t := template.New("text")
	if err != nil {
		fmt.Println(err)
		return false
	}

	t = t.Funcs(template.FuncMap{"genlist": genlist, "iconv": iconv})

	srcfile, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return false
	}

	var buffer [1024 * 1024]byte
	n, rerr := srcfile.Read(buffer[0:])
	if rerr != nil {
		fmt.Println(rerr)
		return false
	}

	t, err = t.Parse(string(buffer[0:n]))
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = t.Execute(file, result.Root)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func main() {

	// src := flag.String("s", `<xml>
	// <return_code>FAIL</return_code>
	// <return_msg>系统繁忙,请稍后再试.</return_msg>
	// <result_code>FAIL</result_code>
	// <err_code>268458547</err_code>
	// <err_code_des>系统繁忙,请稍后再试.</err_code_des>
	// <mch_billno>0010010404201411170000046542</mch_billno>
	// <mch_id>10010404</mch_id>
	// <wxappid>wx6fa7e3bab7e15415</wxappid>
	// <re_openid>onqOjjmM1tad-3ROpncN-yUfa6uI</re_openid>
	// <total_amount>1</total_amount>
	// </xml>`, "The xml string ")

	src := flag.String("s", `<xml>
	<return_code><![CDATA[FAIL]]></return_code>
	<return_msg><![CDATA[系统繁忙,请稍后再试.]]></return_msg>
	<result_code><![CDATA[FAIL]]></result_code>
	<err_code><![CDATA[268458547]]></err_code>
	<err_code_des><![CDATA[系统繁忙,请稍后再试.]]></err_code_des>
	<mch_billno><![CDATA[0010010404201411170000046542]]></mch_billno>
	<mch_id>10010404</mch_id>
	<wxappid><![CDATA[wx6fa7e3bab7e15415]]></wxappid>
	<re_openid><![CDATA[onqOjjmM1tad-3ROpncN-yUfa6uI]]></re_openid>
	<total_amount>1</total_amount>
	</xml>`, "The xml string ")
	name := flag.String("n", "AutoGenerated", "structural name")
	// tag := flag.String("t", "xml", "tag name, many tags can split with '|'")
	// individual := flag.Bool("i", false, "individual each struct")

	flag.Parse()

	// doc := etree.NewDocument()
	// if err := doc.ReadFromString(*src); err != nil {
	// 	panic(err)
	// }
	// fmt.Println(doc)
	// root := doc.SelectElement("xml")
	// fmt.Println("ROOT element:", root.Tag)

	parseIO(strings.NewReader(*src))
	// fmt.Println(xmlResponse)
	fmt.Println("type " + *name + " struct {")
	for k, v := range xmlResponse {
		if v == "xml.Data" {
			fmt.Println(fmt.Sprintf("\t%s string `xml:\"%s,CDATA\"` // %s", FirstUpper(k), k, k))
		} else {
			fmt.Println(fmt.Sprintf("\t%s xml.Name `xml:\"%s,CDATA\"` // %s", FirstUpper(k), k, k))
		}
	}
	fmt.Println("}")

}
