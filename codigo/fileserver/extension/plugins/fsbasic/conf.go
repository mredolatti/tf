package fsbasic

import "fmt"

type Config struct {
	FilePath string
	AuthDBPath string
}

func (c *Config) PopulateFromArgs(args map[string]interface{}) error {
	var ok bool
	if c.FilePath, ok = args["filePath"].(string); !ok {
		return fmt.Errorf("argument 'filePath' missing or incorrect type")
	}

	if c.AuthDBPath, ok = args["authDBPath"].(string); !ok {
		return fmt.Errorf("argument 'authDBPath' missing or incorrect type")
	}

	return nil
}
