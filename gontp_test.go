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
		{"全球授时服务", "pool.ntp.org"},
		{"中国国家授时中心", "ntp.ntsc.ac.cn"},
		{"中科院授时服务", "ntp.ntsc.ac.cn"}, //速度最佳
		{"中国教育网专用", "edu.ntp.org.cn"},
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

func TestUpdateSystemTime(t *testing.T) {
	now, err := GetNtpTime("ntp.ntsc.ac.cn")
	if err != nil {
		t.Error("获取NTP时间错误:", err.Error())
	} else {
		err = UpdateSystemDateTime(now)
		if err != nil {
			t.Error("更新系统时间错误:", err.Error())
		} else {
			t.Log("更新系统时间成功:", now)
		}
	}
}

func TestGetHttpUrl(t *testing.T) {
	now, _, err := GetFromHttpUrl("http://127.0.0.1:8080/api/time")
	if err != nil {
		t.Error("获取NTP时间错误:", err.Error())
	} else {
		t.Log(now)
	}
}
