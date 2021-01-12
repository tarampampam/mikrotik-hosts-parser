package serve

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/pflag"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/env"
)

type flags struct {
	listen struct {
		ip   string
		port uint16
	}
	resourcesDir  string // can be empty
	configPath    string
	cachingEngine string

	// redisDSN allows to setup redis server using single string. Examples:
	//	redis://<user>:<password>@<host>:<port>/<db_number>
	//	unix://<user>:<password>@</path/to/redis.sock>?db=<db_number>
	redisDSN string
}

func (f *flags) init(flagSet *pflag.FlagSet) {
	wd, _ := os.Getwd()

	flagSet.StringVarP(
		&f.listen.ip,
		"listen",
		"l",
		"0.0.0.0",
		fmt.Sprintf("IP address to listen on [$%s]", env.ListenAddr),
	)
	flagSet.Uint16VarP(
		&f.listen.port,
		"port",
		"p",
		8080,
		fmt.Sprintf("TCP port number [$%s]", env.ListenPort),
	)
	flagSet.StringVarP(
		&f.resourcesDir,
		"resources-dir",
		"r",
		filepath.Join(wd, "web"),
		fmt.Sprintf("path to the directory with public assets [$%s]", env.ResourcesDir),
	)
	flagSet.StringVarP(
		&f.configPath,
		"config",
		"c",
		filepath.Join(wd, "configs", "config.yml"),
		fmt.Sprintf("config file path [$%s]", env.ConfigPath),
	)
	flagSet.StringVarP(
		&f.cachingEngine,
		"caching-engine",
		"",
		cachingEngineMemory,
		fmt.Sprintf("caching endine (%s|%s) [$%s]", cachingEngineMemory, cachingEngineRedis, env.CachingEngine),
	)
	flagSet.StringVarP(
		&f.redisDSN,
		"redis-dsn",
		"",
		"redis://127.0.0.1:6379/0",
		fmt.Sprintf("redis server DSN (redis://<user>:<password>@<host>:<port>/<db_number>) [$%s]", env.RedisDSN),
	)
}

func (f *flags) overrideUsingEnv() error {
	if envVar, exists := env.ListenAddr.Lookup(); exists {
		f.listen.ip = envVar
	}

	if envVar, exists := env.ListenPort.Lookup(); exists {
		if p, err := strconv.ParseUint(envVar, 10, 16); err == nil {
			f.listen.port = uint16(p)
		} else {
			return fmt.Errorf("wrong TCP port environment variable [%s] value", envVar)
		}
	}

	if envVar, exists := env.ResourcesDir.Lookup(); exists {
		f.resourcesDir = envVar
	}

	if envVar, exists := env.ConfigPath.Lookup(); exists {
		f.configPath = envVar
	}

	if envVar, exists := env.CachingEngine.Lookup(); exists {
		f.cachingEngine = envVar
	}

	if envVar, exists := env.RedisDSN.Lookup(); exists {
		f.redisDSN = envVar
	}

	return nil
}

func (f *flags) validate() error {
	if net.ParseIP(f.listen.ip) == nil {
		return fmt.Errorf("wrong IP address [%s] for listening", f.listen.ip)
	}

	if f.resourcesDir != "" {
		if info, err := os.Stat(f.resourcesDir); err != nil || !info.Mode().IsDir() {
			return fmt.Errorf("wrong resources directory [%s] path", f.resourcesDir)
		}
	}

	if info, err := os.Stat(f.configPath); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("config file [%s] was not found", f.configPath)
	}

	switch f.cachingEngine {
	case cachingEngineMemory:
	case cachingEngineRedis:
		if _, err := redis.ParseURL(f.redisDSN); err != nil {
			return fmt.Errorf("wrong redis DSN [%s]: %w", f.redisDSN, err)
		}
	default:
		return fmt.Errorf("unsupported caching engine: %s", f.cachingEngine)
	}

	return nil
}