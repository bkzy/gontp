package gontp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"strings"
)

//Based on https://github.com/vladimirvivien/go-ntp-client

// NTP packet format (v3 with optional v4 fields removed)
//
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |LI | VN  |Mode |    Stratum     |     Poll      |  Precision   |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         Root Delay                            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         Root Dispersion                       |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                          Reference ID                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                     Reference Timestamp (64)                  +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                      Origin Timestamp (64)                    +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                      Receive Timestamp (64)                   +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                      Transmit Timestamp (64)                  +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//

// This program implements a trivial NTP client over UDP.
//
// Usage:
// time -e <host endpoint as addr:port>
//

// Get NTP time by ntp server path
// ntpserver like: us.pool.ntp.org
func GetNtpTime(ntpserver string) (now time.Time, er error) {
	const ntpEpochOffset = 2208988800
	type packet struct {
		Settings       uint8  // leap yr indicator, ver number, and mode
		Stratum        uint8  // stratum of local clock
		Poll           int8   // poll exponent
		Precision      int8   // precision exponent
		RootDelay      uint32 // root delay
		RootDispersion uint32 // root dispersion
		ReferenceID    uint32 // reference id
		RefTimeSec     uint32 // reference timestamp sec
		RefTimeFrac    uint32 // reference timestamp fractional
		OrigTimeSec    uint32 // origin time secs
		OrigTimeFrac   uint32 // origin time fractional
		RxTimeSec      uint32 // receive time secs
		RxTimeFrac     uint32 // receive time frac
		TxTimeSec      uint32 // transmit time secs
		TxTimeFrac     uint32 // transmit time frac
	}

	host := ntpserver
	//flag.StringVar(&host, "e", ntpserver, "NTP host")
	///flag.Parse()

	if !strings.Contains(host, ":123") {
		host = ntpserver + ":123"
	}
	// Setup a UDP connection
	conn, err := net.Dial("udp", host)
	if err != nil {
		er = fmt.Errorf("failed to connect: %v", err)
		return
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		er = fmt.Errorf("failed to set deadline: %v", err)
		return
	}
	// configure request settings by specifying the first byte as
	// 00 011 011 (or 0x1B)
	// |  |   +-- client mode (3)
	// |  + ----- version (3)
	// + -------- leap year indicator, 0 no warning
	req := &packet{Settings: 0x1B}

	// send time request
	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		er = fmt.Errorf("failed to send request: %v", err)
		return
	}

	// block to receive server response
	rsp := &packet{}
	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
		er = fmt.Errorf("failed to read server response: %v", err)
		return
	}

	// On POSIX-compliant OS, time is expressed
	// using the Unix time epoch (or secs since year 1970).
	// NTP seconds are counted since 1900 and therefore must
	// be corrected with an epoch offset to convert NTP seconds
	// to Unix time by removing 70 yrs of seconds (1970-1900)
	// or 2208988800 seconds.
	secs := float64(rsp.TxTimeSec) - ntpEpochOffset
	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32 // convert fractional to nanos
	now = time.Unix(int64(secs), nanos)
	return
}

//Update System Date and time
func UpdateSystemDateTime(t time.Time) error {
	dateTime := t.Format("2006-01-02 15:04:05.000000000")
	system := runtime.GOOS

	switch system {
	case "windows":
		dateTime = t.Format("2006-01-02 15:04:05")
		{
			dandt := gstr.Split(dateTime, " ")
			//fmt.Println("日期:", dandt[0], "时间:", dandt[1])
			_, err1 := gproc.ShellExec(`date  ` + dandt[0])
			//fmt.Println("日期转换:", string(GbkToUtf8([]byte(str1))), string(GbkToUtf8([]byte(err1.Error()))))
			if err1 != nil {
				return fmt.Errorf("error updating system date,please start the program as an administrator:%s", err1.Error())
			}
			_, err2 := gproc.ShellExec(`time  ` + dandt[1])
			//fmt.Println("时间转换:", string(GbkToUtf8([]byte(str2))), string(GbkToUtf8([]byte(err2.Error()))))
			if err2 != nil {
				return fmt.Errorf("error updating system time,please start the program as an administrator:%s", err2.Error())
			}
			return nil
		}
	case "linux":
		{
			_, err := gproc.ShellExec(`date -s  "` + dateTime + `"`)
			if err != nil {
				return fmt.Errorf("error updating system time:%s", err.Error())
			}
			return nil
		}
	case "darwin":
		{
			_, err := gproc.ShellExec(`date -s  "` + dateTime + `"`)
			if err != nil {
				return fmt.Errorf("error updating system time:%s", err.Error())
			}
			return nil
		}
	}
	return fmt.Errorf("error updating system time: unsupported operating system < %s >", system)
}

//Convert time string to go time.Time
func TimeParse(s string, loc ...*time.Location) (time.Time, error) {
	location := time.Local
	if len(loc) > 0 {
		for _, v := range loc {
			location = v
		}
	}
	s = strings.ReplaceAll(s, "\"", "")
	if strings.Contains(s, ".") {
		strs := strings.Split(s, ".") //用.切片
		dsec := "."                   //小数点后的秒格式
		tzone := "Z0700"
		havezone := false //是否包含时区
		if len(strs) > 1 {
			dseclen := 0                        //小数部分的长度
			if strings.Contains(strs[1], "+") { //包含时区
				havezone = true
				dotseconds := strings.Split(strs[1], "+") //分割小数的秒和时区
				dseclen = len(dotseconds[0])
				if len(dotseconds) > 1 { //时区切片
					if strings.Contains(dotseconds[1], ":") { //时区切片含有分号
						tzone = "Z07:00"
					}
				}
			} else {
				dseclen = len(strs[1])
			}

			for i := 0; i < dseclen; i++ {
				dsec += "0"
			}
		}
		layout := "2006-01-02 15:04:05"
		if strings.Contains(s, "T") {
			layout = "2006-01-02T15:04:05"
		}
		layout += dsec
		if havezone { //包含时区
			layout += tzone
			t, err := time.Parse(layout, s)
			if err == nil {
				return t, nil
			}
		} else {
			t, err := time.ParseInLocation(layout, s, location)
			if err == nil {
				return t, nil
			}
		}
	} else {
		if strings.Contains(s, "+") { //包含时区
			tzone := "Z0700"
			strs := strings.Split(s, "+") //时区
			if len(strs) > 1 {            //时区切片
				if strings.Contains(strs[1], ":") { //时区切片含有分号
					tzone = "Z07:00"
				}
			}
			layout := "2006-01-02 15:04:05"
			if strings.Contains(s, "T") {
				layout = "2006-01-02T15:04:05"
			}
			layout += tzone
			t, err := time.Parse(layout, s)
			if err == nil {
				return t, nil
			}
		}
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, location)
	if err == nil {
		return t, nil
	}

	t, err = time.ParseInLocation("2006-01-02 15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006年01月02日 15:04:05", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006年01月02日 15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006年01月02日 15点04分05秒", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006年01月02日 15点04分", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-1-2 15:04:05", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-1-2 15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006/01/02 15:04:05", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006/01/02 15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006/1/2 15:04:05", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006/1/2 15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02T15:04:05Z", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02T15:04:05", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02T15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-1-2T15:04:05Z", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-1-2T15:04", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("20060102150405", s, location)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("200601021504", s, location)
	if err == nil {
		return t, nil
	}
	return t, fmt.Errorf("the raw time string is:%s,the error is:%s", s, err.Error())
}

//Get string from http url
func GetFromHttpUrl(urlstr string) (string, int, error) {
	resp, err := http.Get(urlstr) //Get Data
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) //Read data
	if err != nil {
		return "", 0, err
	}

	return string(body), resp.StatusCode, nil
}

func GbkToUtf8(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil
	}
	var bstr []byte
	for _, c := range d {
		if c > 0 {
			bstr = append(bstr, c)
		}
	}
	return bstr
}
