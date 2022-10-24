package model

type GateConfig struct {
	Api struct {
		Authorization struct {
			Need     bool              `json:"need" yaml:"need"`
			Service  string            `json:"service" yaml:"service"`
			Uri      string            `json:"uri" yaml:"uri"`
			Method   string            `json:"method" yaml:"method"`
			Params   map[string]string `json:"params" yaml:"params"`
			Response struct {
				Data string `json:"data" yaml:"data"`
			} `json:"response" yaml:"response"`
		} `json:"authorization" yaml:"authorization"`
		Gates   map[string]serviceConfig `json:"gates" yaml:"gates"`
		Swagger struct {
			Show        bool                         `json:"show" yaml:"show"`
			Description string                       `json:"description" yaml:"description"`
			Title       string                       `json:"title" yaml:"title"`
			Version     string                       `json:"version" yaml:"version"`
			Tags        map[string]map[string]string `json:"tags" yaml:"tags"`
		} `json:"swagger" yaml:"swagger"`
	} `json:"api" yaml:"api"`
}

type serviceConfig struct {
	Service      string       `json:"service" yaml:"service"`
	Prefix       string       `json:"prefix" yaml:"prefix"`
	Allow        []PathConfig `json:"allow" yaml:"allow"`
	Block        []PathConfig `json:"block" yaml:"block"`
	Unauthorized []PathConfig `json:"unauthorized" yaml:"unauthorized"`
}

type PathConfig struct {
	Path   string `json:"path" yaml:"path"`
	Method string `json:"method,omitempty" yaml:"method"`
}
