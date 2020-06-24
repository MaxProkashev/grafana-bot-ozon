package screenshot // загрузка скриншотов из grafana, если не грузится то постится кот

import (
	"io"
	"log"
	"net/http"
)

func DownloadFile(url string, urlCat string) io.Reader {

	resp, err := http.Get(url)
	if err != nil {
		log.Println("problems with grafana screen, do cat")
		resp, _ = http.Get(urlCat)
		return resp.Body
	}

	return resp.Body
}
