package slackpost // пост методом UploadFile в канал и в тред, перевод времени из int64 в string

import (
	"errors"
	"fmt"
	"github.com/nlopes/slack"
	"io"
)

type Slack struct {
	sc *slack.Client
}
type RTM struct {
	RTM *slack.RTM
}
type Parameters struct {
	initialComment string
	filename       string
	channel        string
	resp           io.Reader
	ts             string
}

func (rtm *RTM) ManageConnection() {
	rtm.RTM.ManageConnection()
}
func (rtm *RTM) IncomingEvents() chan slack.RTMEvent {
	return rtm.RTM.IncomingEvents
}

func (rtm *RTM) GetInfo() *slack.Info {
	return rtm.RTM.GetInfo()
}

func New(token string) *Slack {
	return &Slack{slack.New(token)}
}
func NewRTM(api *Slack) *RTM {
	return &RTM{api.sc.NewRTM()}
}
func NewParameters(initialComment string, filename string, channel string, resp io.Reader, ts string) *Parameters {
	return &Parameters{initialComment, filename, channel, resp, ts}
}

func GetThreadTs(SlackChannel string, api *Slack) (ts string, err error) {
	channel, err := api.sc.GetChannelInfo(SlackChannel)
	if err != nil {
		return "", errors.New("can not get info about the channel: " + SlackChannel)
	}
	return channel.Latest.Timestamp, nil
}
func PostUploadFile(parameters *Parameters, api *Slack) error {
	params := slack.FileUploadParameters{
		InitialComment: parameters.initialComment,
		Filename:       parameters.filename,
		Reader:         parameters.resp,
		Channels:       []string{parameters.channel},
	}
	_, err := api.sc.UploadFile(params)
	if err != nil {
		return errors.New("can not send the file to slack")
	}
	return nil
}
func PostUploadFileInThread(parameters *Parameters, api *Slack) error {
	params := slack.FileUploadParameters{
		InitialComment:  parameters.initialComment,
		Filename:        parameters.filename,
		Reader:          parameters.resp,
		Channels:        []string{parameters.channel},
		ThreadTimestamp: parameters.ts,
	}
	_, err := api.sc.UploadFile(params)
	if err != nil {
		return errors.New("can not send the file to slack")
	}
	return nil
}

func CheckEvent(msg slack.RTMEvent) string {
	switch ev := msg.Data.(type) {
	case *slack.ConnectedEvent:
		return "connect"
	case *slack.MessageEvent:
		return ev.Msg.Text
	}
	return ""
}
func PostUploadFilev2(parameters *Parameters, api *Slack) error {
	params := slack.FileUploadParameters{
		Filename: parameters.filename,
		Reader:   parameters.resp,
		Channels: []string{
			"CTBNAGPQE",
		},
	}
	file, err := api.sc.UploadFile(params)
	if err != nil {
		return errors.New("can not send the file to slack")

	}

	apiMe := slack.New("xoxp-3271663252-659813929572-672338229987-574e5c849772305bee93ced52d4ec3ce")

	file, err = apiMe.RevokeFilePublicURL(file.ID)
	if err != nil {
		return err

	}
	file, _, _, err = apiMe.ShareFilePublicURL(file.ID)
	fmt.Println(file)
	if err != nil {
		return err

	}

	options := []slack.MsgOption{slack.MsgOptionAsUser(false)}
	blocks := []slack.Block{
		newTextBlock(parameters.initialComment),
		slack.NewImageBlock(file.URL, parameters.filename, "image", slack.NewTextBlockObject("plain_text", "image", false, false)),
	}
	options = append(options, slack.MsgOptionBlocks(blocks...))

	_, _, err = apiMe.PostMessage(parameters.channel, options...)
	if err != nil {
		fmt.Println(err)
		return errors.New("can not send the file to slack")
	}

	return nil
}
func newTextBlock(text string) slack.Block {
	to := slack.NewTextBlockObject("mrkdwn", text, false, false)

	return slack.NewSectionBlock(to, nil, nil)
}
