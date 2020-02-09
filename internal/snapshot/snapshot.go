package snapshot

import (
	"bytes"
	"compress/zlib"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/micro-kit/micro-common/common"
	"github.com/nfnt/resize"
)

/* 获取摄像头快照 */

var (
	snapshotPath = "snapshot.png"
	cmd          = "ffmpeg"
	baseArgs     = "-f avfoundation -framerate 30 -i 0:0 -f image2 -frames:v 1 -vcodec png " + snapshotPath
)

// GetSnapshot 获取树莓派摄像头图片
func GetSnapshot() ([]byte, error) {
	defer func() {
		// 删除快照
		os.Remove(snapshotPath)
	}()
	// 截图
	log.Println("开始截图")
	args := make([]interface{}, 0)
	argsStr := strings.Split(baseArgs, " ")
	for _, v := range argsStr {
		args = append(args, v)
	}
	session := sh.NewSession()
	session.Command(cmd, args...).Run() // 不处理返回值错误，指令执行完会退出，不影响截图
	// 压缩图片尺寸
	body := ReSizeImg()

	return body, nil
	// return CompressBytes(body), nil // - 压缩效果不好，不压缩了
}

// CompressBytes BytesZlib 压缩字节
func CompressBytes(body []byte) []byte {
	var in bytes.Buffer
	zlibWriter, err := zlib.NewWriterLevel(&in, zlib.BestCompression)
	if err != nil {
		log.Println("创建压缩对象错误", err)
		return []byte{}
	}
	_, err = zlibWriter.Write(body)
	if err != nil {
		zlibWriter.Close()
		log.Println("压缩图片内容错误", err)
		return []byte{}
	}
	zlibWriter.Close()
	log.Println("压缩结果", len(in.Bytes())-len(body))
	return in.Bytes()
}

// ReSizeImg 缩放图片
func ReSizeImg() (body []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("缩放图片错误: v%", err)
			body = []byte{}
		}
		return
	}()
	if isExist, _ := common.PathExists(snapshotPath); isExist == false {
		return []byte{}
	}
	file, err := os.Open(snapshotPath)
	if err != nil {
		log.Println(err)
	}
	img, err := png.Decode(file)
	if err != nil {
		log.Println(err)
	}
	file.Close()
	m := resize.Resize(300, 0, img, resize.Lanczos3)
	out := bytes.NewBuffer(nil)
	png.Encode(out, m)
	return out.Bytes()
}
