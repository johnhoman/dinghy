package generate

type ConfigMap struct {
	Literals []string
	// Files will be read and added to the config file by file name.
	Files []string
	// EnvFiles are files containing key value pairs like
	// <key>=<value>
	EnvFiles []string
}
