package planner

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"errors"
	"sync"
	Logger "gatoor/orca/rewriteTrainer/log"
	"reflect"
	"fmt"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/cloud"
)

var PlannerLogger = Logger.LoggerWithField(Logger.Logger, "module", "planner")
var QueueLogger = Logger.LoggerWithField(PlannerLogger, "object", "queue")

type HostDesiredConfig struct {

}

type LayoutDiff map[base.HostId]map[base.AppName]state_cloud.AppsVersion

type UpdateState string

var Queue PlannerQueue

const (
	STATE_QUEUED = "STATE_QUEUED"
	STATE_APPLYING = "STATE_APPLYING"
	STATE_SUCCESS = "STATE_SUCCESS"
	STATE_FAIL = "STATE_FAIL"
	STATE_UNKNOWN = "STATE_UNKNOWN"
)

type AppsUpdateState struct {
	State UpdateState
	Version state_cloud.AppsVersion
}

type PlannerQueue struct {
	Queue map[base.HostId]map[base.AppName]AppsUpdateState
	lock *sync.Mutex
}

func NewPlannerQueue() *PlannerQueue{
	QueueLogger.Info("Initializing")
	p := &PlannerQueue{}
	p.Queue = make(map[base.HostId]map[base.AppName]AppsUpdateState)
	p.lock = &sync.Mutex{}
	QueueLogger.Info("Initialized")
	return p
}

func (p PlannerQueue) AllEmpty() bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	return len(p.Queue) == 0
}

func (p PlannerQueue) Apply(diff LayoutDiff) {
	QueueLogger.Info("Applying LayoutDiff")
	for hostId, app := range diff {
		for appName, appObj := range app {
			p.Add(hostId, appName, appObj)
		}
	}
}

func (p PlannerQueue) Empty(hostId base.HostId) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		return len(p.Queue[hostId]) == 0
	}
	return true
}

func (p PlannerQueue) Add(hostId base.HostId, appName base.AppName, elem state_cloud.AppsVersion) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; !exists {
		p.Queue[hostId] = make(map[base.AppName]AppsUpdateState)
	}
	if _, exists := p.Queue[hostId][appName]; !exists {
		QueueLogger.Infof("Adding to host '%s' app '%s': '%v'", hostId, appName, elem)
		p.Queue[hostId][appName] = AppsUpdateState{STATE_QUEUED, elem}
	}
}

func (p PlannerQueue) Get(hostId base.HostId) (map[base.AppName]AppsUpdateState, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		if len(p.Queue[hostId]) == 0 {
			return make(map[base.AppName]AppsUpdateState), errors.New(fmt.Sprintf("no element in queue for host %s", hostId))
		}

		elem := p.Queue[hostId]
		return elem, nil
	}
	return make(map[base.AppName]AppsUpdateState), errors.New("failed to get correct AppsVersion")
}

func (p PlannerQueue) GetState(hostId base.HostId, appName base.AppName) (UpdateState, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		if len(p.Queue[hostId]) == 0 {
			return STATE_UNKNOWN, errors.New(fmt.Sprintf("no element in queue for host %s", hostId))
		}

		if elem, ex := p.Queue[hostId][appName]; ex {
			return elem.State, nil
		}
	}
	return STATE_UNKNOWN, errors.New("failed to get correct AppsVersion")
}

func (p PlannerQueue) Remove(hostId base.HostId, appName base.AppName) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		if val, exists := p.Queue[hostId][appName]; exists {
			if val.State == STATE_SUCCESS || val.State == STATE_FAIL {
				QueueLogger.Infof("Removing from host '%s' app '%s'", hostId, appName)
				delete(p.Queue[hostId], appName)
			}
		}
	}
}

func (p PlannerQueue) RemoveApp(appName base.AppName, version base.Version) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for host, apps:= range p.Queue {
		for appN, appObj := range apps {
			if appN == appName && version == appObj.Version.Version {
				QueueLogger.Infof("Removing '%s' - '%s' from Queue of host '%s'", appN, version, host)
				delete(apps, appN)
			}

		}
	}
}


func (p PlannerQueue) SetState(hostId base.HostId, appName base.AppName, state UpdateState) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, exists := p.Queue[hostId]; exists {
		if _, exists := p.Queue[hostId][appName]; !exists {
			return
		}
		elem := p.Queue[hostId][appName]
		elem.State = state
		p.Queue[hostId][appName] = elem
		QueueLogger.Infof("Set state of '%s' '%s' to '%s'", hostId, appName, state)
	}
}

func (p PlannerQueue) Snapshot() map[base.HostId]map[base.AppName]AppsUpdateState{
	p.lock.Lock()
	defer p.lock.Unlock()
	res := make(map[base.HostId]map[base.AppName]AppsUpdateState)
	for k,v := range p.Queue {
		res[k] = v
	}
	QueueLogger.Infof("Created snapshot")
	return res
}

func (p PlannerQueue) RemoveHost(hostId base.HostId) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.Queue, hostId)
	QueueLogger.Infof("Removed host '%s'", hostId)
}


func init() {
	PlannerLogger.Info("Initializing Planner")
	Queue = *NewPlannerQueue()
	PlannerLogger.Info("Initialized Planner")
}

func Diff(master state_cloud.CloudLayout, slave state_cloud.CloudLayout) LayoutDiff {
	diff := LayoutDiff{}
	PlannerLogger.Infof("Generating diff from master: '%v' and slave: '%v'", master, slave)
	for hostId, layoutElem := range master.Layout {
		if _, exists := slave.Layout[hostId]; !exists {
			diff[hostId] = make(map[base.AppName]state_cloud.AppsVersion)
			continue
		}
		tmp := diff[hostId]
		tmp = appsDiff(layoutElem.Apps, slave.Layout[hostId].Apps)
		diff[hostId] = tmp
	}
	PlannerLogger.Infof("Generated diff: '%v'", diff)
	return diff
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
				PlannerLogger.Infof("Got app removal for app '%s'", appName)
				diff[appName] = state_cloud.AppsVersion{versionElem.Version, 0}
			}
		}
	}
	return diff
}

func Plan() {
	PlannerLogger.Info("Stating Plan()")
	doPlan()
	PlannerLogger.Info("Finished Plan()")
}

func getGlobalMissingResources() state_cloud.InstanceResources {
	neededCpu, neededMem, neededNet := getGlobalMinNeeds()
	availableCpu, availableMem, availableNet := getGlobalResources()

	res := state_cloud.InstanceResources{
		TotalCpuResource: state_cloud.CpuResource(int(neededCpu) - int(availableCpu)),
		TotalMemoryResource: state_cloud.MemoryResource(int(neededMem) - int(availableMem)),
		TotalNetworkResource: state_cloud.NetworkResource(int(neededNet) - int(availableNet)),
	}
	return res
}

func InitialPlan() {
	PlannerLogger.Info("Stating initialPlan()")
	neededCpu, neededMem, neededNet := getGlobalMinNeeds()
	availableCpu, availableMem, availableNet := getGlobalResources()

	if int(neededCpu) > int(availableCpu) {
		PlannerLogger.Warnf("Not enough Cpu resources available (needed=%d - available=%d) - spawning new instance TODO", neededCpu, availableCpu)
		cloud.CurrentProvider.SpawnInstances(cloud.CurrentProvider.SuitableInstanceTypes(getGlobalMissingResources()))
		doPlan()
		return
	}
	if int(neededMem) > int(availableMem) {
		PlannerLogger.Warnf("Not enough Memory resources available (needed=%d - available=%d) - spawning new instance TODO", neededMem, availableMem)
		cloud.CurrentProvider.SpawnInstances(cloud.CurrentProvider.SuitableInstanceTypes(getGlobalMissingResources()))
		doPlan()
		return
	}
	if int(neededNet) > int(availableNet) {
		PlannerLogger.Warnf("Not enough Network resources available (needed=%d - available=%d) - spawning new instance TODO", neededNet, availableNet)
		cloud.CurrentProvider.SpawnInstances(cloud.CurrentProvider.SuitableInstanceTypes(getGlobalMissingResources()))
		doPlan()
		return
	}

	doPlan()

	PlannerLogger.Info("Finished initialPlan()")
}


type FailedAssign struct {
	TargetHost base.HostId
	AppName base.AppName
	AppVersion base.Version
	DeploymentCount base.DeploymentCount
}

type MissingAssign struct {
	AppName base.AppName
	AppVersion base.Version
	AppType base.AppType
	DeploymentCount base.DeploymentCount
}

var FailedAssigned []FailedAssign
var failedAssignMutex = &sync.Mutex{}
var MissingAssigned []MissingAssign
var missingssignMutex = &sync.Mutex{}

//TODO fancier optimizations
func doPlan() {
	wipeDesired()
	doPlanInternal()
	handleFailedAssign()
	handleMissingAssign()
	//assignSurplusResources()
}


func appPlanningOrder(allApps map[base.AppName]base.AppConfiguration) ([]base.AppName, []base.AppName) {
	//apps := make([]base.AppName, len(allApps), len(allApps))
	httpApps := make(map[base.AppName]base.AppConfiguration)
	workerApps := make(map[base.AppName]base.AppConfiguration)

	for appName, appObj := range allApps {
		if appObj.Type == base.APP_HTTP {
			httpApps[appName] = appObj
		} else {
			workerApps[appName] = appObj
		}
	}

	httpOrdered := sortByTotalNeeds(httpApps)
	workersOrdered := sortByTotalNeeds(workerApps)

	PlannerLogger.Debugf("Sorted Apps http:%+v, worker:%+v", httpOrdered, workersOrdered)

	//apps = append(httpOrdered, workersOrdered...)
	return httpOrdered, workersOrdered
}



func createChunks(apps []base.AppName) [][]base.AppName{
	const MAX_CONCURRENT = 3
	iter := int(len(apps) / MAX_CONCURRENT)
	if iter < 1 {
		iter = 1
	}
	res := [][]base.AppName{}
	for i:= 0; i < len(apps); i += iter {
		current := iter
		if i + current >= len(apps) {
			current = len(apps) - i
		}
		res = append(res, apps[i:(i+current)])
	}
	return res
}


func doPlanInternal() {
	apps := state_configuration.GlobalConfigurationState.AllAppsLatest()
	httpOrder, workerOrder := appPlanningOrder(apps)

	httpChunks := createChunks(httpOrder)
	for _, chunk := range httpChunks {
		var wg sync.WaitGroup
		wg.Add(len(chunk))
		for _, appName := range chunk {
			appObj := apps[appName]
			PlannerLogger.Infof("Assigning HttpApp '%s' - '%s'. Need to do this %d times", appObj.Name, appObj.Version, appObj.MinDeploymentCount)
			go func () {
				defer wg.Done()
				planHttp(appObj, findHttpHostWithResources, false)
			}()

		}
		wg.Wait()
	}

	for _, appName := range workerOrder {
		appObj := apps[appName]
		PlannerLogger.Infof("Assigning WorkerApp '%s' - '%s'. Need to do this %d times", appObj.Name, appObj.Version, appObj.MinDeploymentCount)
		planWorker(appObj, findHostWithResources, false)

	}
}


func handleMissingAssign() {
	PlannerLogger.Infof("Starting handleMissingAssign for %d elements", len(MissingAssigned))
	PlannerLogger.Error("HANDLE MISSING ASSIGN NOT IMPLEMENTED -- TODO call CloudProvider to spawn instances")
	PlannerLogger.Error("HANDLE MISSING ASSIGN NOT IMPLEMENTED -- TODO call CloudProvider to spawn instances")
	PlannerLogger.Error("HANDLE MISSING ASSIGN NOT IMPLEMENTED -- TODO call CloudProvider to spawn instances")
	cloud.CurrentProvider.SpawnInstance("m1.xlarge")
	PlannerLogger.Infof("handleMissingAssign complete")
}

func handleFailedAssign() {
	PlannerLogger.Infof("Starting handleFailed Assign for %d elements", len(FailedAssigned))
	for _, failed := range FailedAssigned {
		appObj, err := state_configuration.GlobalConfigurationState.GetApp(failed.AppName, failed.AppVersion)
		if err == nil {
			PlannerLogger.Infof("Retrying failed assignment of app '%s', DeploymentCount: %d", failed.AppName, failed.DeploymentCount)
			appObj.MinDeploymentCount = failed.DeploymentCount
			if appObj.Type == base.APP_HTTP {
				planHttp(appObj, findHttpHostWithResources, true)
			} else {
				planWorker(appObj, findHostWithResources, true)
			}
		}
	}
	PlannerLogger.Info("handleFailed Assign complete")
}

func assignSurplusResources() {

}

type HostFinderFunc func (needs state_needs.AppNeeds, app base.AppName, sortedHosts []base.HostId, goodHosts map[base.HostId]bool) base.HostId
type DeploymentCountFunc func (resources state_cloud.InstanceResources, needs state_needs.AppNeeds) base.DeploymentCount

func planApp(appObj base.AppConfiguration, hostFinderFunc HostFinderFunc, deploymentCountFunc DeploymentCountFunc, ignoreFailures bool) bool {
	success := true
	needs, err := state_needs.GlobalAppsNeedState.Get(appObj.Name, appObj.Version)
	if err != nil {
		return false
	}
	var deployed base.DeploymentCount
	deployed = 0
	sortedHosts := sortByAvailableResources()
	goodHosts := state_cloud.GlobalCloudLayout.Current.FindHostsWithApp(appObj.Name)

	for deployed <= appObj.MinDeploymentCount {
		hostId := hostFinderFunc(needs, appObj.Name, sortedHosts, goodHosts)
		if hostId == "" {
			PlannerLogger.Warnf("App '%s' - '%s' could not find suitable host", appObj.Name, appObj.Version)
			success = false
			break
		}
		var depl base.DeploymentCount
		if appObj.Type == base.APP_HTTP {
			depl = 1
		} else {
			resources, err := state_cloud.GlobalAvailableInstances.GetResources(hostId)
			if err != nil {
				break
			}
			depl = deploymentCountFunc(resources, needs)
		}

		if deployed == appObj.MinDeploymentCount {
			PlannerLogger.Infof("Assinged all deployments of App '%s' - '%s'", appObj.Name, appObj.Version)
			return success
		}
		if depl > appObj.MinDeploymentCount - deployed {
			depl = appObj.MinDeploymentCount - deployed
		}
		if !assignAppToHost(hostId, appObj, depl) {
			if !ignoreFailures {
				addFailedAssign(hostId, appObj.Name, appObj.Version, depl)
			} else {
				PlannerLogger.Warnf("Assign of App '%s' - '%s' failed again. Will not try again.", appObj.Name, appObj.Version)
			}
			success = false
		}
		deployed += depl
	}

	if deployed < appObj.MinDeploymentCount {
		PlannerLogger.Warnf("App '%s' - '%s' could not deploy MinDeploymentCount %d, only deployed %d", appObj.Name, appObj.Version, appObj.MinDeploymentCount, deployed)
		addMissingAssign(appObj.Name, appObj.Version, appObj.Type, appObj.MinDeploymentCount - deployed)
		success = false
	}
	return success
}

func planWorker(appObj base.AppConfiguration, hostFinderFunc HostFinderFunc, ignoreFailures bool) bool {
	return planApp(appObj, hostFinderFunc, maxDeploymentOnHost, ignoreFailures)
}

func planHttp(appObj base.AppConfiguration, hostFinderFunc HostFinderFunc, ignoreFailures bool) bool {
	httpDeploymentCountFunc := func(resources state_cloud.InstanceResources, needs state_needs.AppNeeds) base.DeploymentCount {
		return 1
	}
	return planApp(appObj, hostFinderFunc, httpDeploymentCountFunc, ignoreFailures)
}

func maxDeploymentOnHost(resources state_cloud.InstanceResources, needs state_needs.AppNeeds) base.DeploymentCount {
	availCpu := int(resources.TotalCpuResource - resources.UsedCpuResource)
	availMem := int(resources.TotalMemoryResource - resources.UsedMemoryResource)
	availNet := int(resources.TotalNetworkResource - resources.UsedNetworkResource)
	maxCpu := int(availCpu / int(needs.CpuNeeds))
	maxMem := int(availMem / int(needs.MemoryNeeds))
	maxNet := int(availNet / int(needs.NetworkNeeds))
	if maxCpu <= maxMem && maxCpu <= maxNet {
		return base.DeploymentCount(maxCpu)
	}
	if maxMem <= maxCpu && maxMem <= maxNet {
		return base.DeploymentCount(maxMem)
	}
	if maxNet <= maxCpu && maxNet <= maxMem {
		return base.DeploymentCount(maxNet)
	}
	return 0
}


func findHostWithResources(needs state_needs.AppNeeds, app base.AppName, sortedHosts []base.HostId, goodHosts map[base.HostId]bool) base.HostId{
	var backUpHost base.HostId = ""

	for host := range goodHosts {
		if state_cloud.GlobalAvailableInstances.HostHasResourcesForApp(host, needs) {
			PlannerLogger.Infof("Found suitable host '%s'. It already has app '%s' installed", host, app)
			return host
		}
	}

	for _, hostId := range sortedHosts {
		if state_cloud.GlobalAvailableInstances.HostHasResourcesForApp(hostId, needs) {
			PlannerLogger.Infof("Found suitable host '%s'", hostId)
			return hostId
		}
	}
	PlannerLogger.Infof("Found no suitable host which has app '%s' installed. Returning backup host '%s'", app, backUpHost)
	return backUpHost
}

var TotalIter int = 0

func findHttpHostWithResources(needs state_needs.AppNeeds, app base.AppName, sortedHosts []base.HostId, goodHosts map[base.HostId]bool) base.HostId {
	var backUpHost base.HostId = ""

	for host := range goodHosts {
		if state_cloud.GlobalCloudLayout.Desired.HostHasApp(host, app) {
			continue
		}
		if state_cloud.GlobalAvailableInstances.HostHasResourcesForApp(host, needs) {
			PlannerLogger.Infof("Found suitable host '%s'", host)
			return host
		}
	}

	for _, hostId := range sortedHosts {
		TotalIter += 1
		if state_cloud.GlobalCloudLayout.Desired.HostHasApp(hostId, app) {
			continue
		}
		if state_cloud.GlobalAvailableInstances.HostHasResourcesForApp(hostId, needs) {
			PlannerLogger.Infof("Found suitable host '%s'", hostId)
			return hostId
		}
	}
	return backUpHost
}


func wipeDesired() {
	PlannerLogger.Info("Wiping Desired layout")
	FailedAssigned = []FailedAssign{}
	MissingAssigned = []MissingAssign{}
	state_cloud.GlobalCloudLayout.Desired.Wipe()
	for hostId := range state_cloud.GlobalAvailableInstances {
		state_cloud.GlobalCloudLayout.Desired.AddEmptyHost(hostId)
	}
}

func forceAssignToHost(hostId base.HostId, app base.AppConfiguration, count base.DeploymentCount) {

}

func addFailedAssign(host base.HostId, name base.AppName, version base.Version, count base.DeploymentCount) {
	failedAssignMutex.Lock()
	FailedAssigned = append(FailedAssigned, FailedAssign{
		TargetHost: host, AppName: name, AppVersion: version, DeploymentCount: count,
	})
	failedAssignMutex.Unlock()
}

func addMissingAssign(name base.AppName, version base.Version, ty base.AppType, count base.DeploymentCount) {
	missingssignMutex.Lock()
	MissingAssigned = append(MissingAssigned, MissingAssign {
		AppName: name, AppVersion: version, AppType: ty, DeploymentCount: count,
	})
	missingssignMutex.Unlock()
}

func assignAppToHost(hostId base.HostId, app base.AppConfiguration, count base.DeploymentCount) bool {
	PlannerLogger.Infof("Assign '%s' - '%s' to host '%s' %d times", app.Name, app.Version, hostId, count)
	needs, err := state_needs.GlobalAppsNeedState.Get(app.Name, app.Version)
	if err != nil {
		PlannerLogger.Warnf("App '%s' - '%s' on host '%s': GlobalAppsNeedState.Get failed", app.Name, app.Version, hostId)
		addFailedAssign(hostId, app.Name, app.Version, count)
		return false
	}
	deployedNeeds := state_needs.AppNeeds{
		CpuNeeds: state_needs.CpuNeeds(int(needs.CpuNeeds) * int(count)),
		MemoryNeeds: state_needs.MemoryNeeds(int(needs.MemoryNeeds) * int(count)),
		NetworkNeeds: state_needs.NetworkNeeds(int(needs.NetworkNeeds) * int(count)),
	}
	if !state_cloud.GlobalAvailableInstances.HostHasResourcesForApp(hostId, needs) {
		PlannerLogger.Warnf("App '%s' - '%s' on host '%s': Instance resources are insufficient, needed: %+v", app.Name, app.Version, hostId, needs)
		addFailedAssign(hostId, app.Name, app.Version, count)
		return false
	}
	updateInstanceResources(hostId, deployedNeeds)
	state_cloud.GlobalCloudLayout.Desired.AddApp(hostId, app.Name, app.Version, count)
	PlannerLogger.Infof("Assign '%s' - '%s' to host '%s' %d times successful", app.Name, app.Version, hostId, count)
	return true
}

func updateInstanceResources(hostId base.HostId, needs state_needs.AppNeeds)  {
	current, err := state_cloud.GlobalAvailableInstances.GetResources(hostId)
	if err != nil {
		return
	}
	current.UsedCpuResource += state_cloud.CpuResource(needs.CpuNeeds)
	current.UsedMemoryResource += state_cloud.MemoryResource(needs.MemoryNeeds)
	current.UsedNetworkResource += state_cloud.NetworkResource(needs.NetworkNeeds)
	state_cloud.GlobalAvailableInstances.Update(hostId, current)
}


//TODO
func SaveDesired() {
	desired := ""
	timeString, _ := db.GetNow()
	db.Audit.Add(db.BUCKET_AUDIT_DESIRED_LAYOUT, timeString, desired)
}


func getGlobalResources() (state_cloud.CpuResource, state_cloud.MemoryResource, state_cloud.NetworkResource) {
	var totalCpuResources state_cloud.CpuResource
	var totalMemoryResources state_cloud.MemoryResource
	var totalNetworkResources state_cloud.NetworkResource

	for _, resources := range state_cloud.GlobalAvailableInstances {
		totalCpuResources += resources.TotalCpuResource
		totalMemoryResources += resources.TotalMemoryResource
		totalNetworkResources+= resources.TotalNetworkResource
	}
	PlannerLogger.Infof("Total available resources: Cpu: %d, Memory: %d, Network: %d", totalCpuResources, totalMemoryResources, totalNetworkResources)
	return totalCpuResources, totalMemoryResources, totalNetworkResources
}


func getGlobalMinNeeds() (state_needs.CpuNeeds, state_needs.MemoryNeeds, state_needs.NetworkNeeds){
	var totalCpuNeeds state_needs.CpuNeeds
	var totalMemoryNeeds state_needs.MemoryNeeds
	var totalNetworkNeeds state_needs.NetworkNeeds

	for appName, appObj := range state_configuration.GlobalConfigurationState.Apps {
		version := appObj.LatestVersion()
		appNeeds , err := state_needs.GlobalAppsNeedState.Get(appName, version)
		if err != nil {
			PlannerLogger.Warnf("Missing needs for app '%s' - '%s'", appName, version)
			continue
		}
		cpu := int(appObj[version].MinDeploymentCount) * int(appNeeds.CpuNeeds)
		mem := int(appObj[version].MinDeploymentCount) * int(appNeeds.MemoryNeeds)
		net := int(appObj[version].MinDeploymentCount) * int(appNeeds.NetworkNeeds)
		PlannerLogger.Infof("AppMinNeeds for '%s' - '%s': Cpu=%d, Memory=%d, Network=%d", appName, version, cpu, mem, net)
		totalCpuNeeds += state_needs.CpuNeeds(cpu)
		totalMemoryNeeds += state_needs.MemoryNeeds(mem)
		totalNetworkNeeds += state_needs.NetworkNeeds(net)
	}
	PlannerLogger.Infof("GlobalAppMinNeeds: Cpu=%d, Memory=%d, Network=%d", totalCpuNeeds, totalMemoryNeeds, totalNetworkNeeds)
	return totalCpuNeeds, totalMemoryNeeds, totalNetworkNeeds
}


func getGlobalCurrentNeeds() (state_needs.CpuNeeds, state_needs.MemoryNeeds, state_needs.NetworkNeeds) {
	var totalCpuNeeds state_needs.CpuNeeds
	var totalMemoryNeeds state_needs.MemoryNeeds
	var totalNetworkNeeds state_needs.NetworkNeeds

	for hostId, hostObj := range state_cloud.GlobalCloudLayout.Current.Layout {
		for appName, appObj := range hostObj.Apps {
			appNeeds , err := state_needs.GlobalAppsNeedState.Get(appName, appObj.Version)
			if err != nil {
				PlannerLogger.Warnf("Missing needs for app '%s' - '%s'", appName, appNeeds)
				continue
			}
			cpu := int(appObj.DeploymentCount) * int(appNeeds.CpuNeeds)
			mem := int(appObj.DeploymentCount) * int(appNeeds.MemoryNeeds)
			net := int(appObj.DeploymentCount) * int(appNeeds.NetworkNeeds)
			PlannerLogger.Infof("AppNeeds on host '%s': App '%s' - '%s' deployed %d times: Cpu=%d, Memory=%d, Network=%d", hostId, appName, appObj.Version, appObj.DeploymentCount, cpu, mem, net)
			totalCpuNeeds += state_needs.CpuNeeds(cpu)
			totalMemoryNeeds += state_needs.MemoryNeeds(mem)
			totalNetworkNeeds += state_needs.NetworkNeeds(net)
		}
	}
	PlannerLogger.Infof("GlobalAppCurrentNeeds: Cpu=%d, Memory=%d, Network=%d", totalCpuNeeds, totalMemoryNeeds, totalNetworkNeeds)
	return totalCpuNeeds, totalMemoryNeeds, totalNetworkNeeds
}



