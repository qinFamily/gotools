package main

import (
	"io"
	"strings"
	"testing"
)

// 见 https://pay.weixin.qq.com/wiki/doc/api/tools/cash_coupon.php?chapter=13_4&index=3
func Test_parseIO(t *testing.T) {
	type args struct {
		ior io.Reader
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			"cehnggong",
			args{
				strings.NewReader(`<xml>

				<sign><![CDATA[E1EE61A91C8E90F299DE6AE075D60A2D]]></sign>
				
				<mch_billno><![CDATA[0010010404201411170000046545]]></mch_billno>
				
				<mch_id><![CDATA[888]]></mch_id>
				
				<wxappid><![CDATA[wxcbda96de0b165486]]></wxappid>
				
				<send_name><![CDATA[send_name]]></send_name>
				
				<re_openid><![CDATA[onqOjjmM1tad-3ROpncN-yUfa6uI]]></re_openid>
				
				<total_amount><![CDATA[200]]></total_amount>
				
				<total_num><![CDATA[1]]></total_num>
				
				<wishing><![CDATA[恭喜发财]]></wishing>
				
				<client_ip><![CDATA[127.0.0.1]]></client_ip>
				
				<act_name><![CDATA[新年红包]]></act_name>
				
				<remark><![CDATA[新年红包]]></remark>
				
				<scene_id><![CDATA[PRODUCT_2]]></scene_id>
				
				<nonce_str><![CDATA[50780e0cca98c8c8e814883e5caa672e]]></nonce_str>
				
				<risk_info>posttime%3d123123412%26clientversion%3d234134%26mobile%3d122344545%26deviceid%3dIOS</risk_info>
				
				</xml>`),
			},
			true,
		},
		{
			"fanhui",
			args{
				strings.NewReader(`<xml>
		<return_code><![CDATA[SUCCESS]]></return_code>
		
		<return_msg><![CDATA[发放成功.]]></return_msg>
		
		<result_code><![CDATA[SUCCESS]]></result_code>
		
		<err_code><![CDATA[0]]></err_code>
		
		<err_code_des><![CDATA[发放成功.]]></err_code_des>
		
		<mch_billno><![CDATA[0010010404201411170000046545]]></mch_billno>
		
		<mch_id>10010404</mch_id>
		
		<wxappid><![CDATA[wx6fa7e3bab7e15415]]></wxappid>
		
		<re_openid><![CDATA[onqOjjmM1tad-3ROpncN-yUfa6uI]]></re_openid>
		
		<total_amount>1</total_amount>
		
		</xml>`),
			},
			true,
		},
		{
			"fanhuierror",
			args{
				strings.NewReader(`<xml>

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
				
				</xml>`),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseIO(tt.args.ior); got != tt.want {
				t.Errorf("parseIO() = %v, want %v", got, tt.want)
			}
		})
	}
}
