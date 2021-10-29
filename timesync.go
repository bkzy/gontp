package gontp

import (
	"time"

	log "github.com/astaxie/beego/logs"
)

//Time synchronization structure
type TimeSync struct {
	//Synchronization period, minutes. If synchronization is not required, set this parameter to 0
	Period int64
	//time source type,ntp or url. if url,the response must like "2006-01-02T15:04:05.0000Z07:00"
	ServerType string
	//time server path
	//if server type is ntp, it's the ntp server path
	//	the commonly used ntp server path like: pool.ntp.org,ntp1.aliyun.com and so on
	//if server type is url,it's like: http://youip:youport/api/time
	//	the url's response must like "2006-01-02T15:04:05.0000Z07:00"
	Server string
}

//Please call TimeSync.Run in goroutines
func (t *TimeSync) Run() {
	if t.Period == 0 {
		return
	}
	if t.ServerType != "ntp" && t.ServerType != "url" {
		log.Alert("Time synchronization server_type error,it must be 'ntp' or 'url'")
		return
	}
	if len(t.Server) == 0 {
		log.Alert("The time synchronization server is not set")
		return
	}
	log.Info("The time synchronization system starts running")
	var minutes int64
	//log.Debug("配置信息:", t.Period, t.ServerType, t.Server)
	for {
		if minutes == 0 {
			minutes = t.Period
			switch t.ServerType {
			case "ntp":
				//Get time from ntp
				//log.Debug("读取Ntp时间开始")
				now, err := GetNtpTime(t.Server)
				//log.Debug("读取Ntp时间结果:", now, err)
				if err == nil {
					//update time to system
					go func() {
						err := UpdateSystemDateTime(now)
						//log.Debug("更新系统时间结果:", err)
						if err != nil {
							log.Error(err.Error())
						} else {
							log.Info("Proofreading time succeeded")
						}
					}()
				} else {
					log.Error(err.Error())
				}
				//log.Debug("更新系统时间完成")
			case "url":
				//Get time str from url
				tstr, status, err := GetFromHttpUrl(t.Server)
				//log.Debug("URL读取到的结果:", tstr, status, err)
				if err != nil || status != 200 {
					log.Error("failed to get time stamp from url %s", t.Server)
				} else {
					//Parse to time.Time
					now, err := TimeParse(tstr)
					if err != nil {
						log.Error("time format error:%s", err.Error())
					} else {
						//update time to system
						go func() {
							err := UpdateSystemDateTime(now)
							if err != nil {
								log.Error(err.Error())
							} else {
								log.Info("Proofreading time succeeded")
							}
						}()
					}
				}
			default:
				log.Info("The time synchronization server type error:%s", t.ServerType)
			}
		}
		//sleep
		//log.Debug("循环计数器:", minutes)
		time.Sleep(60 * time.Second)
		minutes -= 1
	}
}
