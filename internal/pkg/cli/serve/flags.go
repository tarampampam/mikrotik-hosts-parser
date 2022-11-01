package serve

import (
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/pflag"

	"github.com/tarampampam/mikrotik-hosts-parser/v4/internal/pkg/env"
)

type flags struct {
	listen struct {
		ip   string
		port uint16
	}

	resourcesDir string // can be empty
	configPath   string

	cache struct {
		ttl    string
		engine string
	}

	// redisDSN allows to setup redis server using single string. Examples:
	//	redis://<user>:<password>@<host>:<port>/<db_number>
	//	unix://<user>:<password>@</path/to/redis.sock>?db=<db_number>
	redisDSN string
}

func (f *flags) init(flagSet *pflag.FlagSet) {
	exe, _ := os.Executable()
	exe = path.Dir(exe)

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
		8080, //nolint:gomnd
		fmt.Sprintf("TCP port number [$%s]", env.ListenPort),
	)
	flagSet.StringVarP(
		&f.resourcesDir,
		"resources-dir",
		"r",
		filepath.Join(exe, "web"),
		fmt.Sprintf("path to the directory with public assets [$%s]", env.ResourcesDir),
	)
	flagSet.StringVarP(
		&f.configPath,
		"config",
		"c",
		filepath.Join(exe, "configs", "config.yml"),
		fmt.Sprintf("config file path [$%s]", env.ConfigPath),
	)
	flagSet.StringVarP(
		&f.cache.engine,
		"caching-engine",
		"",
		cachingEngineMemory,
		fmt.Sprintf("caching engine (%s|%s) [$%s]", cachingEngineMemory, cachingEngineRedis, env.CachingEngine),
	)
	flagSet.StringVarP(
		&f.cache.ttl,
		"cache-ttl",
		"",
		"30m",
		fmt.Sprintf("cache entries lifetime (examples: 50s, 1h30m) [$%s]", env.CacheTTL),
	)
	flagSet.StringVarP(
		&f.redisDSN,
		"redis-dsn",
		"",
		"redis://127.0.0.1:6379/0",
		fmt.Sprintf("redis server DSN (format: \"redis://<user>:<password>@<host>:<port>/<db_number>\") [$%s]", env.RedisDSN), //nolint:lll
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
		f.cache.engine = envVar
	}

	if envVar, exists := env.CacheTTL.Lookup(); exists {
		f.cache.ttl = envVar
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

	switch f.cache.engine {
	case cachingEngineMemory:
	case cachingEngineRedis:
		if _, err := redis.ParseURL(f.redisDSN); err != nil {
			return fmt.Errorf("wrong redis DSN [%s]: %w", f.redisDSN, err)
		}
	default:
		return fmt.Errorf("unsupported caching engine: %s", f.cache.engine)
	}

	if _, err := time.ParseDuration(f.cache.ttl); err != nil {
		return fmt.Errorf("wrong cache lifetime [%s] period", f.cache.ttl)
	}

	return nil
}
