package main

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"grafana-bot/pkg/grafana"
	"grafana-bot/pkg/overload"
	screen "grafana-bot/pkg/screenshot" /**/
	"grafana-bot/pkg/slackpost"

	"context"
	_ "grafana-bot/internal/config"

	"gitlab.ozon.ru/platform/scratch"
	_ "gitlab.ozon.ru/platform/scratch/app/pflag"
	"gitlab.ozon.ru/platform/tracer-go/logger"
)

type Config struct {
	Token string `yaml:"token"`

	ImageWidth      string `yaml:"imageWidth"`
	ImageHeight     string `yaml:"imageHeight"`
	OverloadBaseURL string `yaml:"overloadBaseURL"`
	JobInfoURL      string `yaml:"jobInfoURL"`
	OverloadBoard   string `yaml:"overloadBoard"`

	ImageURL1 string `yaml:"imageURL1"`
	ImageURL2 string `yaml:"imageURL2"`

	ViewURL1 string `yaml:"viewURL1"`
	ViewURL2 string `yaml:"viewURL2"`

	FirstCat    string `yaml:"firstCat"`
	NameScreen1 string `yaml:"nameScreen1"`
	NameScreen2 string `yaml:"nameScreen2"`

	Reports []struct {
		Env          string `yaml:"Env"`
		Project      string `yaml:"Project"`
		Name         string `yaml:"Name"`
		SlackChannel string `yaml:"SlackChannel"`
		Panels       []struct {
			Title string `yaml:"Title"`
			URL   string `yaml:"URL"`
		} `yaml:"Panels"`
	} `yaml:"reports"`
}

func main() {

	a, err := scratch.New()
	if err != nil {
		logger.Fatalf(context.Background(), "can't create app: %s", err)
	}

	if err := a.Run(); err != nil {
		logger.Fatalf(context.Background(), "can't run app: %s", err)
	}

	var config Config

	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal("config is not read")
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal("problems with unmarshal config")
	}

	api := slackpost.New(config.Token)
	rtm := slackpost.NewRTM(api)

	go rtm.ManageConnection()
	go func() {
		err := PostReportInSlack(api, config)
		if err != nil {
			log.Println(err)
		}
	}()

	callAt(3, 0, 0, config, api)

	for msg := range rtm.IncomingEvents() {
		Text := slackpost.CheckEvent(msg)
		if Text == "connect" {
			log.Println("connected to Slack")
			continue
		}
		if Text != "" {
			if !strings.Contains(Text, "<@"+rtm.GetInfo().User.ID+">") {
				continue
			}
			log.Println("got a message: " + strings.Split(Text, "<@"+rtm.GetInfo().User.ID+"> ")[1])
			if strings.Contains(Text, "report") {
				go func() {
					err := PostReportInSlack(api, config)
					if err != nil {
						log.Println(err)
					}
				}()
			}
		}
	}

	return
}

func PostReportInSlack(api *slackpost.Slack, config Config) error {
	log.Println("do post in slack")
	for i := range config.Reports {
		infoJob, err := overload.GetInfoLatestOverloadJob(config.OverloadBaseURL, config.Reports[i].Env, config.Reports[i].Project, config.Reports[i].Name)
		if err != nil {
			return err
		}

		startDelFive := strconv.FormatInt(time.Unix(infoJob.GetTestStart(), 0).Add(-time.Minute*5).Unix()*1000, 10)
		stopDelFive := strconv.FormatInt(time.Unix(infoJob.GetTestStop(), 0).Add(time.Minute*5).Unix()*1000, 10)

		startTest := strconv.FormatInt(time.Unix(infoJob.GetTestStart(), 0).Unix()*1000, 10)
		stopTest := strconv.FormatInt(time.Unix(infoJob.GetTestStop(), 0).Unix()*1000, 10)

		urlForDownload := config.OverloadBoard + startTest + "&to=" + stopTest + "&var-test_id=" + strconv.FormatInt(infoJob.GetID(), 10) + "&width=" + config.ImageWidth + "&height=" + config.ImageHeight + "&tz=Europe%2FMoscow"
		resp := screen.DownloadFile(urlForDownload, config.FirstCat)

		infoJS, err := overload.GetInfoJS(strconv.FormatInt(infoJob.GetID(), 10), config.JobInfoURL)
		if err != nil {
			return err
		}

		loadprofile := "linearly growing load from " + infoJS.GetFrom() + " rps to " + infoJS.GetTo() + " rps during " + infoJS.GetDuration()

		moscow, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			log.Println("can not send the file to slack")
			moscow = time.FixedZone("UTC+3", 3*60*60)
		}

		initialComment := "*" + strings.Split(time.Unix(infoJob.GetTestStart(), 0).In(moscow).String(), "+")[0] + " - " + strings.Split(time.Unix(infoJob.GetTestStop(), 0).In(moscow).String(), "+")[0] + "*\n*Description:* " + infoJob.GetDescription() + "\n*Load profile:* " + loadprofile + "\n*AutostopMessage:* " + infoJob.GetAutostopMessage() + "\nhttps://overload.o3.ru/job?id=" + strconv.FormatInt(infoJob.GetID(), 10)

		params := slackpost.NewParameters(initialComment, config.NameScreen1, config.Reports[i].SlackChannel, resp, "")
		err = slackpost.PostUploadFile(params, api)
		if err != nil {
			return err
		}

		tsthread, err := slackpost.GetThreadTs(config.Reports[i].SlackChannel, api)
		if err != nil {
			return err
		}

		for l := range config.Reports[i].Panels {
			infoURL, err := grafana.PreparationURL(config.Reports[i].Panels[l].URL)
			if err != nil {
				continue
			}

			urlForDownload := config.ImageURL1 + infoURL.GetDashboard() + config.ImageURL2 + infoURL.GetService() + "&from=" + startDelFive + "&to=" + stopDelFive + "&panelId=" + infoURL.GetPanelID() + "&width=" + config.ImageWidth + "&height=" + config.ImageHeight + "&tz=Europe%2FMoscow"
			resp := screen.DownloadFile(urlForDownload, config.FirstCat)

			link := config.ViewURL1 + infoURL.GetDashboard() + config.ViewURL2 + infoURL.GetService() + "&from=" + startDelFive + "&to=" + stopDelFive + "&panelId=" + infoURL.GetPanelID() + "&width=" + config.ImageWidth + "&height=" + config.ImageHeight + "&tz=Europe%2FMoscow&fullscreen"

			initialComment := "*<" + link + "|" + config.Reports[i].Panels[l].Title + ">*"

			params := slackpost.NewParameters(initialComment, config.NameScreen1, config.Reports[i].SlackChannel, resp, tsthread)
			err = slackpost.PostUploadFileInThread(params, api)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func callAt(hour, min, sec int, config Config, api *slackpost.Slack) {

	now := time.Now()
	firstCallTime := time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, now.Location())
	if firstCallTime.Before(now) {
		firstCallTime = firstCallTime.Add(time.Hour * 24)
	}

	duration := firstCallTime.Sub(time.Now())

	go func() {
		time.Sleep(duration)
		for {
			err := PostReportInSlack(api, config)
			if err != nil {
				log.Fatal("function broken PostReportInSlack")
			}
			time.Sleep(time.Hour * 24)
		}
	}()
}
