package cloud



type Provider string
type InstanceType string
type ResourceReservoir float32

type Instance struct {
	Type InstanceType
	CpuReservoir ResourceReservoir
	MemoryReservoir ResourceReservoir
	NetworkReservoir ResourceReservoir
}

type MinInstanceCount int
type MaxInstanceCount int
