package program

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/codeskyblue/go-sh"
	piliveserverpb "github.com/micro-kit/microkit-client/proto/piliveserverpb"
	"github.com/micro-kit/pi-live-client/internal/config"
	grpc "google.golang.org/grpc"
)

// Program 程序实体
type Program struct {
	cfg            *config.Config
	pushCmdSession *sh.Session
	rpiSrv         piliveserverpb.PiLiveServerClient
	heartbeatLock  sync.Mutex
	pushLock       sync.Map    // 推流锁，推流时不做截图使用
	pushCh         chan string // 推流状态变化消息 play | stop
}

// New 创建程序实例
func New() (*Program, error) {
	// 初始化配置文件
	cfgChan, err := config.NewConfig("")
	if err != nil {
		log.Println("读取配置错误", err)
		return nil, err
	}
	// 实例对象
	p := &Program{
		cfg:           <-cfgChan,
		heartbeatLock: sync.Mutex{},
		pushLock:      sync.Map{},
		pushCh:        make(chan string, 0),
	}
	p.pushLock.Store("push", false)

	// 创建流服务grpc客户端
	address := fmt.Sprintf("%s:%d", p.cfg.Grpc.Address, p.cfg.Grpc.Port)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Println("连接grpc服务端错误", err)
		return nil, err
	}
	rpiSrv := piliveserverpb.NewPiLiveServerClient(conn)
	if err != nil {
		log.Println("创建流服务grpc客户端错误", err)
		return nil, err
	}
	p.rpiSrv = rpiSrv

	return p, nil
}

// Run 启动程序
func (p *Program) Run() {
	// 后期需要改成需要时推流
	go p.PushByCh()
	// 开启心跳
	go func() {
		// 这里做循环处理，如果函数返回，则再次启动
		for {
			time.Sleep(3 * time.Second)
			p.Heartbeat()
			log.Println("心跳遇到错误，重新启动")
		}
	}()
	// 定时截图
	go p.UploadTask()
}

// Stop 程序结束要做的事
func (p *Program) Stop() {
	p.StopPush()
}
