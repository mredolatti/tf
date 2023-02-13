package main

import (
	"fmt"

	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts"
	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	"github.com/mredolatti/tf/codigo/fileserver/extension/plugins/fsbasic"
)

type Plugin struct {
	auth      *fsbasic.Authorization
	files     *fsbasic.Files
	filesmeta *fsbasic.FilesMetadata
}

// GetAuthorization implements apiv1.Plugin
func (p *Plugin) GetAuthorization() apiv1.Authorization {
	return p.auth
}

// GetFileMetadataStorage implements apiv1.Plugin
func (p *Plugin) GetFileMetadataStorage() apiv1.FilesMetadata {
	return p.filesmeta
}

// GetFileStorage implements apiv1.Plugin
func (p *Plugin) GetFileStorage() apiv1.Files {
	return p.files
}

func Create(args map[string]interface{}) (apiv1.Plugin, error) {
	var p Plugin
	var cfg fsbasic.Config
	if err := cfg.PopulateFromArgs(args); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	var err error
	if p.auth, err = fsbasic.NewAuthz(cfg.AuthDBPath); err != nil {
		return nil, fmt.Errorf("error setting up authorization db: %w", err)
	}
	if p.files, err = fsbasic.NewFiles(cfg.FilePath); err != nil {
		return nil, fmt.Errorf("error setting up file repository: %w", err)
	}

	if p.filesmeta, err = fsbasic.NewFilesMetadata(cfg.FilePath); err != nil {
		return nil, fmt.Errorf("error setting up file meta repository: %w", err)
	}

	return &p, nil
}

func APIVersion() contracts.Version {
	return apiv1.V
}

var _ apiv1.Plugin = (*Plugin)(nil)
