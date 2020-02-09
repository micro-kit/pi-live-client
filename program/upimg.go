package program

import (
	"context"
	"log"
	"time"

	piliveserverpb "github.com/micro-kit/microkit-client/proto/piliveserverpb"
	"github.com/micro-kit/pi-live-client/internal/snapshot"
)

/* 定时上传截图 */

// var out bytes.Buffer
// r, _ := zlib.NewReader(&in)
// io.Copy(&out, r)
// fmt.Println(out.String())

// UploadTask 定时任务
func (p *Program) UploadTask() {
	time.Sleep(3 * time.Second)
	p.uploadSnapshot() // 启动就拍快照
	t := time.NewTicker(30 * time.Second)
	for {
		<-t.C
		log.Println("开始准备上传快照")
		p.uploadSnapshot()
	}
}

// 上传摄像头快照
func (p *Program) uploadSnapshot() {
	// 推流中不截图
	isPush, ok := p.pushLock.Load("push")
	if ok {
		if isPush.(bool) == true {
			return
		}
	} else {
		return
	}
	// 获取截图内容
	body, err := snapshot.GetSnapshot()
	if len(body) == 0 {
		log.Println("快照大小为0，不发送数据")
		return
	}
	// 发送数据到服务端
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()
	resp, err := p.rpiSrv.UploadSnapshot(ctx, &piliveserverpb.UploadSnapshotRequest{
		Appname: p.cfg.Live.AppName,
		LiveId:  p.cfg.Live.LiveId,
		Body:    body,
	})
	if err != nil {
		log.Println("上传摄像头快照错误", err)
		return
	}
	log.Println("上传快照结果", resp.Status)
}
