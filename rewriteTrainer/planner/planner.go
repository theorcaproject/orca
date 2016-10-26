package planner

import (
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"errors"
	"sync"
	Logger "gatoor/orca/rewriteTrainer/log"
	"reflect"
	"fmt"
)

var PlannerLogger = Logger.LoggerWithField(Logger.Logger, "module", "planner")

type HostDesiredConfig struct {

}

type UpdateState string

const (
	STATE_QUEUED = "STATE_QUEUED"
	STATE_APPLYING = "STATE_APPLYING"
	STATE_SUCCESS = "STATE_SUCCESS"
	STATE_FAIL = "STATE_FAIL"
)

type AppsUpdateState struct {
	State UpdateState
	Version state_cloud.AppsVersion
}

type PlannerQueue struct {
	Queue map[base.HostId][]AppsUpdateState
	lock *sync.Mutex
}

func NewPlannerQueue() *PlannerQueue{
	p := &PlannerQueue{}
	p.Queue = make(map[base.HostId][]AppsUpdateState)
	p.lock = &sync.Mutex{}
	return p
}

func (p PlannerQueue) AllEmpty() bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	return len(p.Queue) == 0
}

func (p PlannerQueue) Empty(hostId base.HostId) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		return len(p.Queue[hostId]) == 0
	}
	return true
}

func (p PlannerQueue) Append(hostId base.HostId, elem state_cloud.AppsVersion) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; !exists {
		p.Queue[hostId] = []AppsUpdateState{}
	}
	p.Queue[hostId] = append(p.Queue[hostId], AppsUpdateState{STATE_QUEUED, elem})
}

func (p PlannerQueue) Pop(hostId base.HostId) (state_cloud.AppsVersion, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		if len(p.Queue[hostId]) == 0 {
			return state_cloud.AppsVersion{}, errors.New(fmt.Sprintf("no element in queue for host %s", hostId))
		}
		elem := p.Queue[hostId][0]
		if elem.State == STATE_SUCCESS || elem.State == STATE_FAIL {
			p.Queue[hostId] = append(p.Queue[hostId][:0], p.Queue[hostId][0+1:]...)
		}
		return elem.Version, nil
	}
	return state_cloud.AppsVersion{}, errors.New("failed to get correct AppsVersion")
}

func (p PlannerQueue) SetState(hostId base.HostId, state UpdateState) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		if len(p.Queue[hostId]) == 0 {
			return
		}
		elem := p.Queue[hostId][0]
		elem.State = state
		p.Queue[hostId][0] = elem
	}
}

func (p PlannerQueue) Snapshot() map[base.HostId][]AppsUpdateState{
	p.lock.Lock()
	defer p.lock.Unlock()
	res := make(map[base.HostId][]AppsUpdateState)
	for k,v := range p.Queue {
		res[k] = v
	}
	return res
}

func (p PlannerQueue) RemoveHost(hostId base.HostId) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.Queue, hostId)
}


var UpdateQueue PlannerQueue
var LayoutQueue PlannerQueue

func Init() {
	UpdateQueue = *NewPlannerQueue()
	LayoutQueue = *NewPlannerQueue()
}

func Config(host base.HostId) HostDesiredConfig {
	// get latest element from layout queue
	return HostDesiredConfig{}
}

func UpdateLayoutQueue(globalLayout state_cloud.CloudLayoutAll) {
	PlannerLogger.Info("Start updating LayoutQueue")
	changeCount := 0

	PlannerLogger.Infof("Done updating LayoutQueue. Changed %d Elements", changeCount)
}

func Diff(master state_cloud.CloudLayout, slave state_cloud.CloudLayout) {
	diff := state_cloud.CloudLayout{}
	for hostId, layoutElem := range master {
		if _, exists := slave[hostId]; !exists {
			diff[hostId] = layoutElem
			continue
		}
		tmp := diff[hostId]
		tmp.Apps = appsDiff(layoutElem.Apps, slave[hostId].Apps)
		diff[hostId] = tmp
	}
}

func appsDiff(master map[base.AppName]state_cloud.AppsVersion, slave map[base.AppName]state_cloud.AppsVersion) map[base.AppName]state_cloud.AppsVersion {
	diff := make(map[base.AppName]state_cloud.AppsVersion)
	for appName, versionElem := range master {
		if !reflect.DeepEqual(versionElem, slave[appName]) {
			diff[appName] = master[appName]
		}
	}
	//handle removal of an app from host:
	for appName, versionElem := range slave {
		if _, exists := master[appName]; !exists {
			if !reflect.DeepEqual(versionElem, master[appName]) {
				diff[appName] = state_cloud.AppsVersion{versionElem.Version, 0}
			}
		}
	}
	return diff
}

func CreateVision() {
	// take state.needs and state.currentCloudLayout and come up with a new vision
	// check Desired to make sure the changes are not too radical
}

func CreateDesired() {
	// take Path and change vision with some delay
}

func UpdateCurrent() {
	//called with hostInfo from instance push api call
	//or instance died event from cloud provider
}

func Schedule() {
	
}




