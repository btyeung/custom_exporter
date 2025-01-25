package config

import (
	"io/ioutil"
	"os/user"
	"runtime"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
)

const (
	Namespace = "custom"
	Exporter  = "exporter"
)

type CredentialsItem struct {
	Name      string `yaml:"name"`
	Collector string `yaml:"type"`
	User      string `yaml:"user,omitempty"`
	Dsn       string `yaml:"dsn,omitempty"`
	Uri       string `yaml:"uri,omitempty"`
	Path      string `yaml:"path,omitempty"`
}

type MetricsItem struct {
	Name       string
	Commands   []string
	Credential CredentialsItem
	Mapping    []string
	Separator  string
	Value_name string
	Value_type prometheus.ValueType
}

type MetricsItemYaml struct {
	Name       string   `yaml:"name"`
	Commands   []string `yaml:"commands"`
	Credential string   `yaml:"credential"`
	Mapping    []string `yaml:"mapping"`
	Separator  string   `yaml:"separator,omitempty"`
	Value_name string   `yaml:"value_name,omitempty"`
	Value_type string   `yaml:"value_type"`
}

type ConfigYaml struct {
	Credentials []CredentialsItem `yaml:"credentials"`
	Metrics     []MetricsItemYaml `yaml:"metrics"`
}

type Config struct {
	Metrics map[string]MetricsItem
}

type CredentialsUser struct {
	user.User
}

func NewConfig(configFile string) (*Config, error) {
	contentFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	ymlCnf := ConfigYaml{}
	if err = yaml.Unmarshal(contentFile, &ymlCnf); err != nil {
		return nil, err
	}

	myCnf := new(Config)
	myCnf.metricsList(ymlCnf)

	return myCnf, nil
}

func (c Config) credentialsList(yaml ConfigYaml) map[string]CredentialsItem {
	result := make(map[string]CredentialsItem)
	for _, v := range yaml.Credentials {
		result[v.Name] = CredentialsItem{
			Name:      v.Name,
			Collector: v.Collector,
			User:      v.User,
			Dsn:       v.Dsn,
			Path:      v.Path,
			Uri:       v.Uri,
		}
	}
	return result
}

func (e Config) ValueType(Value_type string) prometheus.ValueType {
	switch Value_type {
	case "COUNTER":
		return prometheus.CounterValue
	case "GAUGE":
		return prometheus.GaugeValue
	default:
		return prometheus.UntypedValue
	}
}

func (c *Config) metricsList(yaml ConfigYaml) {
	result := make(map[string]MetricsItem)
	credentials := c.credentialsList(yaml)

	for _, v := range yaml.Metrics {
		if cred, ok := credentials[v.Credential]; ok {
			result[v.Name] = MetricsItem{
				Name:       v.Name,
				Commands:   v.Commands,
				Credential: cred,
				Mapping:    v.Mapping,
				Separator:  v.Separator,
				Value_name: v.Value_name,
				Value_type: c.ValueType(v.Value_type),
			}
		} else {
			log.Fatalf("error credential, collector type not found : %s", v.Credential)
		}
	}

	c.Metrics = result
}

func (m MetricsItem) SeparatorValue() string {
	if len(m.Separator) < 1 {
		return " "
	}
	return m.Separator
}

func (m MetricsItem) CredentialUser() *CredentialsUser {
	// On Windows, we don't support running as different users
	if runtime.GOOS == "windows" {
		return currentUser()
	}

	usr := strings.TrimSpace(m.Credential.User)
	if len(usr) == 0 {
		return currentUser()
	}

	if myUser, err := user.LookupId(usr); err == nil {
		return &CredentialsUser{User: *myUser}
	}

	if myUser, err := user.Lookup(usr); err == nil {
		return &CredentialsUser{User: *myUser}
	}

	return currentUser()
}

func currentUser() *CredentialsUser {
	myUser, err := user.Current()
	if err != nil {
		log.Fatalf("Error retrieving current system user: %s", err.Error())
	}
	return &CredentialsUser{User: *myUser}
}

func (c CredentialsUser) UidInt() uint32 {
	// On Windows, we don't use UIDs
	if runtime.GOOS == "windows" {
		return 0
	}
	
	if uid, err := strconv.ParseUint(c.Uid, 10, 32); err == nil {
		return uint32(uid)
	}
	return 0
}

func (c CredentialsUser) GidInt() uint32 {
	// On Windows, we don't use GIDs
	if runtime.GOOS == "windows" {
		return 0
	}
	
	if gid, err := strconv.ParseUint(c.Gid, 10, 32); err == nil {
		return uint32(gid)
	}
	return 0
}
