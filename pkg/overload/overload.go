package overload // все что связано с запросами к overload последним job-ам их json

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type collectionsInfo struct {
	Collections []struct {
		Env        string `json"env"`
		Project    string `json"project"`
		Name       string `json:"name"`
		LatestJobs []struct {
			Id              int64  `json"id"`
			TestStart       int64  `json"testStart"`
			TestStop        int64  `json"testStop"`
			Author          string `json"author"`
			Description     string `json"description"`
			AutostopMessage string `json"autostopMessage"`
		} `json:"latestJobs"`
	} `json"collections"`
}

type jobInfo struct {
	Jobs []struct {
		Config string `json"config"`
	} `json"jobs"`
}

type infoJS struct {
	duration string
	from     string
	to       string
}

func (info *infoJS) GetDuration() string {
	return info.duration
}
func (info *infoJS) GetFrom() string {
	return info.from
}
func (info *infoJS) GetTo() string {
	return info.to
}

type infoLatestOverloadJob struct {
	id              int64
	testStart       int64
	testStop        int64
	description     string
	autostopMessage string
}

func (info *infoLatestOverloadJob) GetID() int64 {
	return info.id
}
func (info *infoLatestOverloadJob) GetTestStart() int64 {
	return info.testStart
}
func (info *infoLatestOverloadJob) GetTestStop() int64 {
	return info.testStop
}
func (info *infoLatestOverloadJob) GetDescription() string {
	return info.description
}
func (info *infoLatestOverloadJob) GetAutostopMessage() string {
	return info.autostopMessage
}

func GetInfoLatestOverloadJob(OverloadBaseURL string, env string, project string, name string) (*infoLatestOverloadJob, error) {

	var overloadinfo collectionsInfo

	req, err := http.NewRequest("GET", OverloadBaseURL+"env="+env+"&project="+project, nil)
	if err != nil {
		return &infoLatestOverloadJob{0, 0, 0, "", ""}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if resp.Status != "200 OK" || err != nil {
		return &infoLatestOverloadJob{0, 0, 0, "", ""}, err
	}

	defer resp.Body.Close()
	reader, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &infoLatestOverloadJob{0, 0, 0, "", ""}, err
	}
	resp.Body.Close()
	err = json.Unmarshal(reader, &overloadinfo)
	if err != nil {
		return &infoLatestOverloadJob{0, 0, 0, "", ""}, err
	}

	for k := range overloadinfo.Collections {
		if overloadinfo.Collections[k].Name != name {
			continue
		}
		if time.Unix(overloadinfo.Collections[k].LatestJobs[0].TestStart, 0).Day() != time.Now().Day() && time.Unix(overloadinfo.Collections[k].LatestJobs[0].TestStart, 0).Day() != time.Now().Day()-1 {
			return &infoLatestOverloadJob{0, 0, 0, "", ""}, errors.New("no new overloads in the last two days")
		}
		return &infoLatestOverloadJob{overloadinfo.Collections[k].LatestJobs[0].Id, overloadinfo.Collections[k].LatestJobs[0].TestStart, overloadinfo.Collections[k].LatestJobs[0].TestStop, overloadinfo.Collections[k].LatestJobs[0].Description, overloadinfo.Collections[k].LatestJobs[0].AutostopMessage}, nil
	}

	return &infoLatestOverloadJob{0, 0, 0, "", ""}, errors.New("the cycle does not work")
}

func GetInfoJS(ID string, JobInfoURL string) (*infoJS, error) {

	var info jobInfo

	req, err := http.NewRequest("GET", JobInfoURL+ID, nil)
	if err != nil {
		return &infoJS{"", "", ""}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if resp.Status != "200 OK" || err != nil {
		return &infoJS{"", "", ""}, err
	}

	defer resp.Body.Close()
	reader, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &infoJS{"", "", ""}, err
	}
	resp.Body.Close()
	err = json.Unmarshal(reader, &info)
	if err != nil {
		return &infoJS{"", "", ""}, err
	}

	allInfo := info.Jobs[0].Config

	infoJS := new(infoJS)

	i := strings.Split(allInfo, "- {duration: ")[1]
	infoJS.duration = strings.Split(i, ",")[0]

	j := strings.Split(i, "from: ")[1]
	infoJS.from = strings.Split(j, ",")[0]

	j = strings.Split(i, "to: ")[1]
	infoJS.to = strings.Split(j, ",")[0]

	return infoJS, nil
}
