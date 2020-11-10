package slack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/johntdyer/slack-go"
	"github.com/sirupsen/logrus"
)

// Sender ...
type Sender struct {
	isDebug bool
	API     string
}

// Hook ...
type Hook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []logrus.Level
	HookURL        string
	IconURL        string
	Channel        string
	IconEmoji      string
	Username       string
	Env            string
	Asynchronous   bool
	Extra          map[string]interface{}
	Disabled       bool
}

// Levels ...
func (sh *Hook) Levels() []logrus.Level {
	if sh.AcceptedLevels == nil {
		return AllLevels
	}
	return sh.AcceptedLevels
}

// Fire ...
func (sh *Hook) Fire(e *logrus.Entry) error {
	if sh.Disabled {
		return nil
	}

	color := ""
	switch e.Level {
	case logrus.DebugLevel:
		color = "#9B30FF"
	case logrus.InfoLevel:
		color = "good"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = "danger"
	default:
		color = "warning"
	}

	msg := &slack.Message{
		Username:  sh.Username,
		Channel:   sh.Channel,
		IconEmoji: sh.IconEmoji,
		IconUrl:   sh.IconURL,
	}

	attach := msg.NewAttachment()

	newEntry := sh.newEntry(e)
	// If there are fields we need to render them at attachments
	if len(newEntry.Data) > 0 {

		// Add a header above field data
		attach.Text = fmt.Sprintf("Application Env: `%s` :ghost:", sh.Env)

		items := make([]*struct {
			Keys  string
			Value string
		}, 0)

		for k, v := range newEntry.Data {
			items = append(items, &struct {
				Keys  string
				Value string
			}{
				Keys:  k,
				Value: fmt.Sprint(v),
			})
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].Keys > items[j].Keys
		})

		output := ""
		for _, i := range items {
			output += fmt.Sprintf("%s: `%s`, ", i.Keys, i.Value)
		}
		output = strings.TrimRight(output, ", ")

		slackField := &slack.Field{}
		slackField.Title = "Trace"
		slackField.Value = output
		if len(slackField.Value) <= 20 {
			slackField.Short = true
		}
		attach.AddField(slackField)

		attach.Pretext = fmt.Sprintf(":icecream: %s `%s`", newEntry.Message, e.Level.String())
	} else {
		attach.Text = fmt.Sprintf(":icecraem: %s `%s`", newEntry.Message, e.Level.String())
	}
	attach.Fallback = fmt.Sprintf(":icecream: %s `%s`", newEntry.Message, e.Level.String())
	attach.Color = color

	c := slack.NewClient(sh.HookURL)

	if sh.Asynchronous {
		go c.SendMessage(msg)
		return nil
	}

	return c.SendMessage(msg)
}

func (sh *Hook) newEntry(entry *logrus.Entry) *logrus.Entry {
	data := map[string]interface{}{}

	for k, v := range sh.Extra {
		data[k] = v
	}
	for k, v := range entry.Data {
		data[k] = v
	}

	newEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	return newEntry
}
