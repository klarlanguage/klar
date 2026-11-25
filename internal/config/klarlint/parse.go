package klarlint

import (
	"github.com/klarlanguage/klar/internal/config"
	"github.com/klarlanguage/klar/pkg/klon"
)

func Parse(path string) (conf *File, warn []*klon.Error, err error) {
	conf = &File{}
	if warn, err = config.ReadFromFile(path, &conf, Context); err != nil {
		return conf, warn, err
	}
	return conf, warn, nil
}

// Default returns the default lint configuration.
func Default() *File { return &File{Config: &Config{}} }

var Context = &klon.Context{
	Enums: map[string]map[string]any{
		"DiagnosticLevel": {
			"true":  DefaultLevel,
			"false": SuppressLevel,
			"warn":  WarningLevel,
			"error": ErrorLevel,
		},
	},
}
