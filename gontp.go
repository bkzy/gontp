package gontp

import (
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"

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

//ntpserver like:us.pool.ntp.org
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

//Update System Date
func UpdateSystemDate(t time.Time) error {
	dateTime := t.Format("2006-01-02 15:04:05.00000000")
	system := runtime.GOOS
	switch system {
	case "windows":
		{
			dandt := gstr.Split(dateTime, " ")
			_, err1 := gproc.ShellExec(`date  ` + dandt[0])
			_, err2 := gproc.ShellExec(`time  ` + dandt[1])
			if err1 != nil && err2 != nil {
				return fmt.Errorf("error updating system time: please start the program as an administrator")
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
