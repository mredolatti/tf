package filemanager

import (
	"encoding/json"
	"fmt"

	authzBasic "github.com/mredolatti/tf/codigo/fileserver/authz/basic"
	"github.com/mredolatti/tf/codigo/fileserver/storage/basic"
)

func Setup(pluginPath string, pluginConf string) (Interface, error) {
	if pluginPath == "" {
		return fallback()
	}

	var pluginParams map[string]interface{}
	if err := json.Unmarshal([]byte(pluginConf), &pluginParams); err != nil {
		return nil, fmt.Errorf("error parsing plugin config JSON: %w", err)
	}

	return fromPlugin(pluginPath, pluginParams)
}

func fallback() (Interface, error) {
	return New(
		basic.NewInMemoryFileStore(),
		basic.NewInMemoryFileMetadataStore(),
		authzBasic.NewInMemoryAuthz(),
	), nil
}
