package gontp

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/gogf/gf/os/gproc"
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
	str, _, err := GetFromHttpUrl("http://127.0.0.1:8080/api/time")
	if err != nil {
		t.Error("获取NTP时间错误:", err.Error())
	} else {
		now, err := TimeParse(str)
		if err != nil {
			t.Error("时间转换错误:", err.Error())
		} else {
			t.Log(now)
		}
	}
}

func TestExceCmd(t *testing.T) {
	c := exec.Command("cmd", "/C", "ping", "127.0.0.1")
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if str, err := c.Output(); err != nil {
		cmderr := string(GbkToUtf8(stderr.Bytes()))
		t.Error(err.Error(), cmderr)
	} else {
		msg := string(GbkToUtf8([]byte(str)))
		t.Log(msg)
	}
}

func TestShellExec(t *testing.T) {
	//var str string
	str, err := gproc.ShellExec("ping 127.0.0.1")
	if err != nil {
		t.Error(string(GbkToUtf8([]byte(str))), string(GbkToUtf8([]byte(err.Error()))))
	} else {
		t.Log(string(GbkToUtf8([]byte(str))))
	}
}

func TestTimeSync(t *testing.T) {
	tsync := new(TimeSync)
	tsync.Period = 2
	tsync.ServerType = "ntp"
	tsync.Server = "ntp.ntsc.ac.cn"
	go tsync.Run()
	for i := 0; i < 100; i++ {
		time.Sleep(1 * time.Second)
	}
}

func TestTimeParse(t *testing.T) {
	tests := []struct {
		timestr string
	}{
		{"2021-10-28T15:02:07.5622783+08:00"},
		{"2021-10-28T15:02:07.5622+08:00"},
		{"2021-10-28T15:02:07.56+08:00"},
		{"2021-10-28 15:02:07.56"},
		{"2021-10-28 15:02:07+0800"},
		{"2021-10-29T18:32:10.26367+08:00"},
	}
	for _, tt := range tests {
		tm, err := TimeParse(tt.timestr)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(tm)
		}
	}
}
