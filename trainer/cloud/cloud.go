package cloud

import "gatoor/orca/trainer/state"

type CloudProvider struct {

}

func (cloud* CloudProvider) ActionChange(change *state.ChangeServer){

}

func (cloud *CloudProvider) HasChanges() bool {
	return false;
}

func (cloud *CloudProvider) GetAllChanges() []*state.ChangeServer {
	return []*state.ChangeServer{}
}

func (cloud* CloudProvider) RemoveChange(changeId string){

}
