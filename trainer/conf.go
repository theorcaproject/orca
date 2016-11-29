package main

import "gatoor/orca/base"

type TrainerJsonConfiguration struct {
	Port int
}

type AppJsonConfiguration struct {
	Name base.AppName
	Version base.Version
	Type base.AppType
	InstallCommands []base.OsCommand
	QueryStateCommand base.OsCommand
	RemoveCommand base.OsCommand
	Needs AppNeeds
}

type HostJsonConfiguration struct {

}

type HabitatJsonConfiguration struct {
	Version base.Version
	InstallCommands []base.OsCommand
}

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

type CloudJsonConfiguration struct {
	InstanceType InstanceType
	MinInstanceCount MinInstanceCount
	MaxInstanceCount MaxInstanceCount
}

type JsonConfiguration struct {
	Trainer TrainerJsonConfiguration
	Habitats []HabitatJsonConfiguration
	Apps []AppJsonConfiguration
}
