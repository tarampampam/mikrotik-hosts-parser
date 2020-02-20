package serve

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"text/tabwriter"

	"github.com/a8m/envsubst"
	"gopkg.in/yaml.v2"
)

type (
	Listen struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	}

	Resources struct {
		DirPath      string `yaml:"dir"`
		IndexName    string `yaml:"index_name"`
		Error404Name string `yaml:"error_404_name"`
	}

	Source struct {
		URI              string `yaml:"uri"`
		Name             string `yaml:"name"`
		Description      string `yaml:"description"`
		EnabledByDefault bool   `yaml:"enabled"`
		RecordsCount     int    `yaml:"count"` // approximate quantity
	}

	Redirect struct {
		Address string `yaml:"address"`
	}

	Excludes struct {
		Hosts []string `yaml:"hosts"`
	}

	Cache struct {
		File        CacheFiles `yaml:"files"`
		LifetimeSec int        `yaml:"lifetime_sec"`
	}

	CacheFiles struct {
		DirPath string `yaml:"dir"`
	}

	RouterScript struct {
		Redirect      Redirect `yaml:"redirect"`
		Exclude       Excludes `yaml:"exclude"`
		Comment       string   `yaml:"comment"`
		MaxSources    int      `yaml:"max_sources"`
		MaxSourceSize int      `yaml:"max_source_size"` // in bytes
	}
)

type Settings struct {
	Listen       Listen       `yaml:"listen"`
	Resources    Resources    `yaml:"resources"`
	Sources      []Source     `yaml:"sources"`
	Cache        Cache        `yaml:"cache"`
	RouterScript RouterScript `yaml:"router_script"`
}

// Creates new settings instance using YAML file.
func FromYamlFile(filename string, expandEnv bool) (*Settings, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromYaml(bytes, expandEnv)
}

// Creates new settings instance using YAML-structured content.
func FromYaml(in []byte, expandEnv bool) (*Settings, error) {
	settings := &Settings{}

	if expandEnv {
		parsed, err := envsubst.Bytes(in)
		if err != nil {
			return nil, err
		}
		in = parsed
	}

	if err := yaml.Unmarshal(in, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// PrintInfo about most important settings values into writer.
func (s *Settings) PrintInfo(out io.Writer) error {
	w := tabwriter.NewWriter(out, 2, 8, 1, '\t', tabwriter.AlignRight)
	defer func(w *tabwriter.Writer) { _ = w.Flush() }(w)

	lines := [][]interface{}{
		{"Listen address", s.Listen.Address},
		{"Listen port", s.Listen.Port},
		{"Resources dir", s.Resources.DirPath},
		{"Index file name", s.Resources.IndexName},
		{"Error 404 file name", s.Resources.Error404Name},
		{"Sources count", len(s.Sources)},
		{"Cache lifetime (sec)", s.Cache.LifetimeSec},
		{"Cache files directory", s.Cache.File.DirPath},
		{"Max sources count", s.RouterScript.MaxSources},
		{"Max source response size (bytes)", s.RouterScript.MaxSourceSize},
	}

	for _, line := range lines {
		_, err := fmt.Fprintf(w, strings.Repeat("%v\t|\t", len(line))+"\n", line...)
		if err != nil {
			return err
		}
	}

	return nil
}
