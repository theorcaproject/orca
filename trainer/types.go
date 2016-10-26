package main

import (
	"gatoor/orca/base"
	"fmt"
)

type AppVersion struct {
	Version base.Version
	Configuration base.AppConfiguration
	Needs AppNeeds
	Stats AppStats
}

type App struct {
	AppName base.AppName
	Versions map[base.Version]AppVersion
}

type Needs struct {
	val float32
}

func (m Needs) Get() (Needs){
	return m
}

func (m Needs) Set(n float32) {
	if n < 0 {
		Logger.Warn(fmt.Sprintf("Needs set too low. Val was %d", n))
		m.val = 0.0
	} else if n > 1 {
		Logger.Warn(fmt.Sprintf("Needs set too high. Val was %d", n))
		m.val = 1.0
	} else {
		m.val = n
	}
}

type MemoryNeeds Needs
type CPUNeeds Needs
type NetworkNeeds Needs

type AppNeeds struct {
	MemoryNeeds MemoryNeeds
	CPUNeeds CPUNeeds
	NetworkNeeds NetworkNeeds
}

type AppStats struct {
	MemoryUsage float32
	CpuUsage float32
	NetworkUsage float32
}


type Habitat struct {
	Version base.Version
	Configuration base.HabitatConfiguration
}

type Host struct {
	Version base.Version
	Configuration base.HostConfiguration
}




