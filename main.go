package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

var server_port int = 9999 // 对端的端口
var server_ip string = ""  // 对端的ip

var logger *zap.Logger

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./udpClient.log",
		MaxSize:    1, // 单位是MB
		MaxBackups: 5,
		MaxAge:     180,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func initLog() error {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger = zap.New(core, zap.AddCaller())
	return nil
}

func parseArgs() {
	flag.IntVar(&server_port, "server_port", 9999, "对端的端口")
	flag.StringVar(&server_ip, "server_ip", "127.0.0.1", "对端的ip")

	flag.Parse()
	logger.Info("parse args finish", zap.Int("server_port", server_port), zap.String("server_ip", server_ip))
}

func main() {
	initLog()

	parseArgs()

	server_add := fmt.Sprintf("%s:%d", server_ip, server_port)
	udpServer, err := net.ResolveUDPAddr("udp", server_add)
	if err != nil {
		logger.Error("ResolveUDPAddr fail", zap.Error(err))
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, udpServer)
	if err != nil {
		logger.Error("DialUDP fail", zap.Error(err))
		os.Exit(1)
	}

	//close the connection
	defer conn.Close()
	var count int64 = 0

	ticker := time.NewTicker(5 * time.Second) // 创建一个5秒的定时器
	go func() {
		for {
			select {
			case <-ticker.C:
				count = count + 1
				send_msg := fmt.Sprintf("%d hello!", count)
				logger.Debug("this send msg:", zap.String("send_msg", send_msg))

				_, err = conn.Write([]byte(send_msg))
				if err != nil {
					logger.Error("Write data failed", zap.Error(err))
					continue
				}

				// buffer to get data
				received := make([]byte, 4096)

				recv_len, recv_err := conn.Read(received)
				if recv_err != nil {
					logger.Error("Read data failed", zap.Error(recv_err))
					continue
				}
				logger.Debug("recv response msg:", zap.Int("recv_len", recv_len), zap.String("received", string(received[:recv_len])))
			}
		}
	}()

	// 主协程永不退出
	for {
		time.Sleep(10 * time.Second)
	}
}
