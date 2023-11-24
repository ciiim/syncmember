package syncmember

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

var (
	//Interval
	FastHeartBeatInterval    = 300 * time.Millisecond
	NormalHeartBeatInterval  = 500 * time.Millisecond
	SlowHeartBeatInterval    = 1 * time.Second
	DefaultHeartBeatInterval = NormalHeartBeatInterval

	FastPushPullInterval    = 10 * time.Millisecond
	NormalPushPullInterval  = 20 * time.Second
	SlowPushPullInterval    = 30 * time.Second
	DefaultPushPullInterval = NormalPushPullInterval

	FastGossipInterval    = 300 * time.Millisecond
	NormalGossipInterval  = 500 * time.Millisecond
	SlowGossipInterval    = 1 * time.Second
	DefaultGossipInterval = NormalGossipInterval
	///

	//HeartBeat and Goosip
	DefaultFanout        = 3
	DefaultUDPBufferSize = 2048
	DefaultPushPullNums  = 1

	//Net
	DefaultTCPTimeout = 5 * time.Second
	BindAllIP         = net.ParseIP("0.0.0.0")
	BindLoopBackIP    = net.ParseIP("127.0.0.1")
	DefaultBindPort   = 9632

	LocalAdvertiseIP     = net.ParseIP("127.0.0.1")
	DefaultAdvertisePort = DefaultBindPort
	//

	//Log
	DefaultLogLevel  = slog.LevelInfo
	OpenLogDetail    = true
	CloseLogDetail   = false
	DefaultLogWriter = os.Stderr
	//
)

type Config struct {
	BindIP        net.IP
	BindPort      int
	AdvertiseIP   net.IP
	AdvertisePort int

	HeartBeatInterval time.Duration
	PushPullInterval  time.Duration

	TCPTimeout time.Duration

	LogLevel  slog.Level
	LogWriter io.Writer
	LogDetail bool

	Fanout       int
	PushPullNums int

	UDPBufferSize int
}

var (
	DefaultConfig = func() *Config {
		return &Config{
			BindIP:   BindAllIP,
			BindPort: DefaultBindPort,

			AdvertisePort: DefaultAdvertisePort,

			HeartBeatInterval: DefaultHeartBeatInterval,
			PushPullInterval:  DefaultPushPullInterval,

			TCPTimeout: DefaultTCPTimeout,

			LogDetail: CloseLogDetail,
			LogLevel:  DefaultLogLevel,
			LogWriter: DefaultLogWriter,

			Fanout: DefaultFanout,

			UDPBufferSize: DefaultUDPBufferSize,
		}

	}

	DebugConfig = &Config{
		BindIP:   BindAllIP,
		BindPort: DefaultBindPort,

		AdvertisePort: DefaultAdvertisePort,

		HeartBeatInterval: DefaultHeartBeatInterval,
		PushPullInterval:  DefaultPushPullInterval,

		TCPTimeout: DefaultTCPTimeout,

		LogDetail: OpenLogDetail,
		LogLevel:  slog.LevelDebug,
		LogWriter: DefaultLogWriter,

		Fanout:       DefaultFanout,
		PushPullNums: DefaultPushPullNums,

		UDPBufferSize: DefaultUDPBufferSize,
	}
)

func initLogger(config *Config) *slog.Logger {
	logHandler := slog.NewJSONHandler(config.LogWriter, &slog.HandlerOptions{
		AddSource: config.LogDetail,
		Level:     config.LogLevel,
	})
	logger := slog.New(logHandler)
	return logger
}

func (s *SyncMember) readConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if config.AdvertisePort == 0 {
		return fmt.Errorf("invalid advertise port")
	}
	if config.BindIP == nil || config.BindPort == 0 {
		return fmt.Errorf("invalid bind ip or port")
	}

	if config.AdvertiseIP == nil {
		config.AdvertiseIP = GetHostIP()
	}

	s.host = ResolveAddr(fmt.Sprintf("%s:%d", config.AdvertiseIP.String(), config.AdvertisePort))

	s.logger = initLogger(config)
	s.heartBeatTicker = time.NewTicker(config.HeartBeatInterval)
	s.pushPullTicker = time.NewTicker(config.PushPullInterval)

	return nil
}

func (c *Config) SetAdvertiserIP(ip string) *Config {
	c.AdvertiseIP = net.ParseIP(ip)
	return c
}

func (c *Config) SetPort(port int) *Config {
	c.AdvertisePort = port
	c.BindPort = port
	return c
}

func (c *Config) SetBindIP(ip string) *Config {
	c.BindIP = net.ParseIP(ip)
	return c
}

func (c *Config) SetHeartBeatInterval(d time.Duration) *Config {
	c.HeartBeatInterval = d
	return c
}

func (c *Config) SetPushPullInterval(d time.Duration) *Config {
	c.PushPullInterval = d
	return c
}

func (c *Config) SetTCPTimeout(d time.Duration) *Config {
	c.TCPTimeout = d
	return c
}

func (c *Config) SetLogLevel(level slog.Level) *Config {
	c.LogLevel = level
	return c
}

func (c *Config) SetLogWriter(w io.Writer) *Config {
	c.LogWriter = w
	return c
}

func (c *Config) OpenLogDetail(detail bool) *Config {
	c.LogDetail = detail
	return c
}

func (c *Config) SetFanout(fanout int) *Config {
	c.Fanout = fanout
	return c
}

func (c *Config) SetUDPBufferSize(size int) *Config {
	c.UDPBufferSize = size
	return c
}
