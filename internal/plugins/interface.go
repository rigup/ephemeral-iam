package plugins

// EIAMPlugin is the interface that is exposed to external plugins to implement.
type EIAMPlugin interface {
	GetInfo() (name, desc, version string, err error)
	Run([]string) error
}
