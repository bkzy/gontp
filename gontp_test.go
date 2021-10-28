package gontp

import (
	"fmt"
	"testing"
)

func TestGetNtpTime(t *testing.T) {
	tests := []struct {
		name   string
		server string
	}{
		{"中国国家授时中心", "ntp.ntsc.ac.cn"},
		{"中科院授时服务", "ntp.ntsc.ac.cn"},
		{"中国教育网专用", "edu.ntp.org.cn"},
		{"阿里云公共授时服务,编号1~7", "time1.aliyun.com"},
		{"阿里云公共授时服务,编号1~7", "ntp1.aliyun.com"},
		{"腾讯公共授时服务,编号1~5", "time1.cloud.tencent.com"},
	}
	for _, tt := range tests {
		now, err := GetNtpTime(tt.server)
		if err != nil {
			t.Error(fmt.Sprintf("%20s", tt.name), err)
		} else {
			t.Log(fmt.Sprintf("%20s", tt.name), now)
		}
	}
}
