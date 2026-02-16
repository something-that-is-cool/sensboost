package embeddable

import (
	"encoding/json"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Marshaller interface {
	Unmarshal(data []byte, dest any) error
}

// YAML ...
var YAML = yamlMarshaller{}

// yamlMarshaller implements Marshaller for YAML data.
type yamlMarshaller struct{}

// Unmarshal ...
func (yamlMarshaller) Unmarshal(data []byte, dest any) error {
	return yaml.Unmarshal(data, dest)
}

// TOML ...
var TOML = tomlMarshaller{}

// tomlMarshaller implements Marshaller for TOML data.
type tomlMarshaller struct{}

// Unmarshal ...
func (tomlMarshaller) Unmarshal(data []byte, dest any) error {
	return toml.Unmarshal(data, dest)
}

// JSON ...
var JSON = jsonMarshaller{}

// jsonMarshaller implements Marshaller for JSON data.
type jsonMarshaller struct{}

// Unmarshal ...
func (jsonMarshaller) Unmarshal(data []byte, dest any) error {
	return json.Unmarshal(data, dest)
}
