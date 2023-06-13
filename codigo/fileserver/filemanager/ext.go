package filemanager

import (
	"fmt"
	"plugin"

	"github.com/mredolatti/tf/codigo/fileserver/authz"
	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts"
	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	v1adapters "github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1/adapters"
)

func fromPlugin(fn string, params map[string]interface{}) (Interface, error) {

	pl, err := plugin.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("error opening plugin file: %w", err)
	}

	version, err := getVersion(pl)
	if err != nil {
		return nil, fmt.Errorf("error looking up plugin version: %w", err)
	}

	switch version {
	case apiv1.V:
		return buildFromV1Plugin(pl, params)
	default:
		return nil, fmt.Errorf("unknown plugin version '%d'", version)
	}
}

func getVersion(p *plugin.Plugin) (contracts.Version, error) {

	var invalid contracts.Version
	symbol, err := p.Lookup(contracts.APIVersionFuncName)
	if err != nil {
		return invalid, fmt.Errorf("error retrieving API Version symbol: %w", err)
	}

	vfunc, ok := symbol.(contracts.APIVersionFunc)
	if !ok {
		return invalid, fmt.Errorf("APIVersion symbol found but has invalid type: '%T'", symbol)
	}

	return vfunc(), nil
}

func buildFromV1Plugin(pl *plugin.Plugin, params map[string]interface{}) (Interface, error) {

	symbol, err := pl.Lookup(apiv1.CreateFuncName)
	if err != nil {
		return nil, fmt.Errorf("error looking up init method in plugin: %w", err)
	}

	create, ok := symbol.(apiv1.CreateFunc)
	if !ok {
		return nil, fmt.Errorf("Create func found but has invalid type: '%T'", symbol)
	}

	plug, err := create(params)
	if err != nil {
		return nil, fmt.Errorf("error invoking plugin creation method: %w", err)
	}

	metaStore := v1adapters.NewFilesMetaWrapper(plug.GetFileMetadataStorage())
	fileStore := v1adapters.NewFilesWrapper(plug.GetFileStorage())
	authorization := v1adapters.NewAuthWrapper(plug.GetAuthorization())

	authorization.Grant("martin.redolatti", authz.OperationRead, authz.AnyObject)
	authorization.Grant("martin.redolatti", authz.OperationWrite, authz.AnyObject)
	authorization.Grant("martin.redolatti", authz.OperationCreate, authz.AnyObject)
	// end stuff to remove

	return New(fileStore, metaStore, authorization), nil
}
