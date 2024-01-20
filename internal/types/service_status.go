package types

type ServiceStatus struct {
	// Name is the name of the service.
	Name string
	// Version is the version of the service.
	Version string

	// Mode is the mode of the service.
	Mode string
}

func (s ServiceStatus) String() string {
	return s.Name + " _" + s.Version + "_ (" + s.Mode + ")"
}
