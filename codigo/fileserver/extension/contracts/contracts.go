package contracts

type Version uint32

type APIVersionFunc = func() Version

const (
	APIVersionFuncName string = "APIVersion"
)
