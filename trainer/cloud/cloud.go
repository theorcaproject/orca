package cloud

import (
	"gatoor/orca/trainer/model"
)

type CloudProvider struct {

}

func (cloud* CloudProvider) ActionChange(change *model.ChangeServer){

}

func (cloud *CloudProvider) HasChanges() bool {
	return false;
}

func (cloud *CloudProvider) GetAllChanges() []*model.ChangeServer {
	return []*model.ChangeServer{}
}

func (cloud* CloudProvider) RemoveChange(changeId string){

}
