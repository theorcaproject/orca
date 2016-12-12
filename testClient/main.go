package testClient

import (
	"bytes"
	"encoding/json"
	"gatoor/orca/base"
	"net/http"
	"io/ioutil"
	"fmt"
)


const (
	TRAINER_URL = "http://localhost:5000/push"
)


func main() {
	CallTrainer()
}

func AppInfo() []base.AppInfo {
	return []base.AppInfo {
		{
			Type: "",
			Name: "",
			Version: 0,
			Status: "",
			Id: "",
		},
	}
}

func HostInfo() base.HostInfo{
	return base.HostInfo{
		HostId: "",
		IpAddr: "",
		OsInfo: nil,
		Apps: AppInfo(),
	}
}

func Metrics() base.MetricsWrapper {
	return base.MetricsWrapper{make(map[string]base.HostStats), base.AppMetrics{}}
}


func CallTrainer() {
	fmt.Println("Calling Trainer...")

	wrapper := base.TrainerPushWrapper{HostInfo(), Metrics()}

	b := new(bytes.Buffer)
	jsonErr := json.NewEncoder(b).Encode(wrapper)
	if jsonErr != nil {
		fmt.Printf("Could not encode Metrics: %+v. Sending without metrics.", jsonErr)
		wrapper.Stats = base.MetricsWrapper{}
		json.NewEncoder(b).Encode(wrapper)
	}
	res, err := http.Post(TRAINER_URL, "application/json; charset=utf-8", b)
	if err != nil {
		fmt.Errorf("Could not send data to trainer: %+v", err)
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Errorf("Could not read reponse from trainer: %+v", err)
		} else {
			var config base.PushConfiguration
			if err := json.Unmarshal(body, &config); err != nil {
				fmt.Errorf("Could not read json from trainer: %+v", err)
			}
			js, _ := json.MarshalIndent(config, "", "    ")
			fmt.Println("--------")
			fmt.Println("--------")
			fmt.Println(js)
			fmt.Println("")
			fmt.Println("--------")
			fmt.Println("--------")
		}
	}
}