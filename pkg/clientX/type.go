package clientX

type Service struct {
	IDC  string
	Bala Balancer[*Addr]
	*ServiceConfig
}

type ServiceConfig struct {
	Name string `toml:"Name"`

	ConnTimeOut  int64 `toml:"ConnTimeOut"`
	WriteTimeOut int64 `toml:"WriteTimeOut"`
	ReadTimeOut  int64 `toml:"ReadTimeOut"`

	Retry int `toml:"Retry"`

	Protocol  string `toml:"Protocol"`
	Converter string `toml:"Converter"`
	Strategy  string `toml:"Strategy"`

	Reuse bool `toml:"Reuse"`

	Resource *ResourcesConfig `toml:"Resource"`
}

type ResourcesConfig struct {
	Manual ManualConfig `toml:"Manual,omitempty"`
	// TODO zookeeper
	// TODO etcd
}

type ManualConfig map[string][]*Addr

type Addr struct {
	IP   string `toml:"IP,omitempty"`
	Port string `toml:"Port,omitempty"`
	Host string `toml:"Host,omitempty"`
	IDC  string `toml:"Idc,omitempty"`
}
