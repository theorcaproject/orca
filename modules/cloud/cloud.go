package cloud

type CloudLayout map[string]CloudLayoutElement

type CloudLayoutElement struct {
	HabitatVersion string
	AppsVersion map[string]string
}

type CloudProvider interface {
	NewInstance() string
}
