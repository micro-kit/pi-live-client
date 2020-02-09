package program

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/micro-kit/micro-common/crypto"
	piliveserverpb "github.com/micro-kit/microkit-client/proto/piliveserverpb"
)

/* 心跳，定时发送流播放信息到服务端 */

// Heartbeat 心跳
func (p *Program) Heartbeat() {
	p.heartbeatLock.Lock()
	defer p.heartbeatLock.Unlock()
	// 请求心跳
	heartbeatLiveClient, err := p.rpiSrv.HeartbeatLive(context.Background())
	if err != nil {
		log.Println("建立心跳连接错误", err)
		return
	}
	// 建立两个任务，读和写
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			data, err := heartbeatLiveClient.Recv()
			if err == io.EOF {
				log.Println("服务端端口流心跳流")
				return
			}
			if err != nil {
				log.Println("读取流错误", err)
				return
			}
			log.Println("服务端响应 Status:", data.Status, "推流状态", data.Push)
			// 心跳播放状态处理
			if data.Push != "" {
				p.pushCh <- data.Push
			}
		}
	}()
	// 定时发送
	go func() {
		defer wg.Done()
		defer heartbeatLiveClient.CloseSend()
		t := time.NewTicker(10 * time.Second)
		for {
			<-t.C
			// 计算签名
			now := strings.ToUpper(fmt.Sprintf("%x", time.Now().Unix()))
			signature := crypto.Md5(now + p.cfg.Live.LiveId)
			log.Println("发送心跳")
			// 发送心跳
			err := heartbeatLiveClient.Send(&piliveserverpb.HeartbeatLiveRequest{
				Appname:   p.cfg.Live.AppName,
				LiveId:    p.cfg.Live.LiveId,
				Time:      now,
				Name:      p.cfg.Live.Name,
				Signature: signature,
			})
			if err != nil {
				log.Println("发送心跳错误", err)
				return
			}
		}
	}()
	wg.Wait()
}
