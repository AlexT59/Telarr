package types

type ServiceStatus struct {
	// Name is the name of the service.
	Name string
	// Version is the version of the service.
	Version string

	// Running is true if the service is running.
	Running bool
}

func (s ServiceStatus) String() string {
	str := "*" + s.Name + "* _v" + s.Version + "_"

	if s.Running {
		str += " (running  ✅)"
	} else {
		str += " (not running  ❌)"
	}

	return str
}
