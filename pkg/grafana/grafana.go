package grafana // препарация ссылки grafana

import (
	"errors"
	"strings"
)

type infoURL struct {
	dashboard string
	service   string
	panelId   string
}

func (info *infoURL) GetDashboard() string {
	return info.dashboard
}
func (info *infoURL) GetService() string {
	return info.service
}
func (info *infoURL) GetPanelID() string {
	return info.panelId
}

func PreparationURL(url string) (*infoURL, error) { // специально для service-overview для других дашбордов нужно убрать service но пока так

	info := new(infoURL)

	if (!strings.Contains(url, "/d/") && !strings.Contains(url, "/render/d-solo/")) || !strings.Contains(url, "service") || !strings.Contains(url, "panelId") {
		return info, errors.New("incorrect URL")
	}

	if strings.Contains(url, "/d/") {
		info.dashboard = strings.Split(strings.Split(url, "/d/")[1], "/")[0]
		if info.dashboard == "" {
			return info, errors.New("incorrect dashboard field in URL")
		}
	}
	if strings.Contains(url, "/render/d-solo/") {
		info.dashboard = strings.Split(strings.Split(url, "/render/d-solo/")[1], "/")[0]
		if info.dashboard == "" {
			return info, errors.New("incorrect dashboard field in URL")
		}
	}

	info.service = strings.Split(strings.Split(url, "service=")[1], "&")[0]
	if info.service == "" {
		return info, errors.New("incorrect service field in URL")
	}

	info.panelId = strings.Split(strings.Split(url, "panelId=")[1], "&")[0]
	if info.panelId == "" {
		return info, errors.New("incorrect panel ID field in URL")
	}

	info.dashboard += "/service-overview"

	return info, nil
}
