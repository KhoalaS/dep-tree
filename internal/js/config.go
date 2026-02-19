package js

type Config struct {
	TsConfigPaths    bool   `yaml:"tsConfigPaths"`
	Workspaces       bool   `yaml:"workspaces"`
	TsConfigFileName string `yaml:"tsConfigFilename"`
}
