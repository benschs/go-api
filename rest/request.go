package rest

type Request struct {
	Name    string            `json:"name" yaml:"name"`
	Func    string            `json:"func" yaml:"func"`
	Method  string            `json:"method" yaml:"method"`
	URI     string            `json:"uri" yaml:"uri"`
	Headers []string          `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body    BodyType          `json:"body,omitempty" yaml:"body,omitempty"`
	Params  map[string]string `json:"params,omitempty" yaml:"params,omitempty"` // URL Params
	Query   []string          `json:"query,omitempty" yaml:"query,omitempty"`   // Query params
}

type BodyType struct {
	IsJSON         bool   `yaml:"isJSON,omitempty"`
	JSONStructName string `yaml:"jsonStructName,omitempty"`

	IsMultipart bool            `yaml:"isMultipart,omitempty"`
	Forms       []MultipartForm `yaml:"forms,omitempty"`
}

type MultipartForm struct {
	Name       string `yaml:"name,omitempty"`
	IsFile     bool   `yaml:"isFile,omitempty"`
	StructName string `yaml:"structName,omitempty"`
}
