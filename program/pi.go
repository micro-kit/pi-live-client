package program

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/codeskyblue/go-sh"
)

/* 树莓派推流控制 */

const (
	// cmd = "raspivid"
	// baseArgs = "-o - -t 0 -vf -hf -fps 10 -b 500000 -w 640 -h 360 | ffmpeg -re -ar 44100 -ac 2 -acodec pcm_s16le -f s16le -ac 2 -i /dev/zero -f h264 -i - -vcodec copy -acodec aac -ab 128k -g 50 -strict experimental -f flv"
	cmd      = "ffmpeg"
	baseArgs = "-f avfoundation -video_size 640x480 -framerate 30 -i 0:0 -vcodec libx264 -preset veryfast -f flv"
) // ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/movie

const (
	// PlayStatusPlay 播放
	PlayStatusPlay = "play"
	// PlayStatusStop 停止
	PlayStatusStop = "stop"
)

// PushByCh 根据消息推流
func (p *Program) PushByCh() {
	for {
		log.Println("处理推流状态变化消息")
		select {
		case push := <-p.pushCh:
			if push == PlayStatusStop {
				if p.pushCmdSession != nil {
					p.pushCmdSession.Kill(os.Kill)
				}
			} else if push == PlayStatusPlay {
				go p.Push()
			} else {
				log.Println("收到未知推流状态", push)
			}
		}
	}
}

// Push 推流
func (p *Program) Push() {
	// 已经再推流情况
	if isPush, ok := p.pushLock.Load("push"); ok && isPush.(bool) == true {
		log.Println("已经在推流了")
		return
	}
	// 推流时锁一下，不再截图
	p.pushLock.Store("push", true)
	defer func() {
		p.pushLock.Store("push", false)
		// 推流结束，截图
		time.Sleep(10 * time.Second)
		p.uploadSnapshot()
	}()
	// 拆散参数
	args := make([]interface{}, 0)
	argsStr := strings.Split(baseArgs, " ")
	for _, v := range argsStr {
		args = append(args, v)
	}
	// 追加流地址
	liveAddr := fmt.Sprintf("%s/%s/%s", p.cfg.Live.BaseURL, p.cfg.Live.AppName, p.cfg.Live.LiveId)
	args = append(args, liveAddr)
	// 执行命令
	log.Println("开始推流: ", liveAddr)
	session := sh.NewSession()
	p.pushCmdSession = session
	err := p.pushCmdSession.Command(cmd, args...).Run()
	if err != nil {
		log.Println("推流错误", err)
	}
}

// StopPush 停止推流
func (p *Program) StopPush() {
	if p.pushCmdSession != nil {
		log.Println("停止推流")
		p.pushCmdSession.Kill(syscall.SIGKILL)
	}
}
