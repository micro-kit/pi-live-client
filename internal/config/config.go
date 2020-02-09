package config

import (
	"os"

	"github.com/micro-kit/pi-live-client/internal/common"
	"github.com/naoina/toml"
)

// Config 配置文件
type Config struct {
	Debug bool         `toml:"debug"`
	Live  LiveConfig   `toml:"live"`
	Grpc  ServerConfig `toml:"grpc"`
}

// LiveConfig 推流配置
type LiveConfig struct {
	Name    string `toml:"name"`     // 显示名
	BaseURL string `toml:"base_url"` // 基础推流地址
	AppName string `toml:"app_name"` // 推流app名
	LiveId  string `toml:"live_id"`  // 流id
}

// ServerConfig grpc 服务配置
type ServerConfig struct {
	Address string `toml:"address"` // 服务端地址
	Port    int    `toml:"port"`    // 端口
}

// NewConfig 初始化一个server配置文件对象
func NewConfig(path string) (cfgChan chan *Config, err error) {
	if path == "" {
		path = common.GetRootDir() + "config/cfg.toml"
	}
	cfgChan = make(chan *Config, 0)
	// 读取配置文件
	cfg, err := readConfFile(path)
	if err != nil {
		return
	}
	go watcher(cfgChan, path)
	go func() {
		cfgChan <- cfg
	}()
	return
}

// ReadConfFile 读取配置文件
func readConfFile(path string) (cfg *Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg = new(Config)
	if err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return
}
