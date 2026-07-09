package plugins

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("json").
		Description("Pretty print JSON for logging").
		Category("Tools").
		Handle(handleJsonLog)
}
func handleJsonLog(m *telegram.NewMessage) error {
	var jsonString []byte
	defer m.Delete()
	if !m.IsReply() {
		if strings.Contains(m.Args(), "-s") {
			jsonString, _ = json.MarshalIndent(m.Sender, "", "  ")
		} else if strings.Contains(m.Args(), "-m") {
			jsonString, _ = json.MarshalIndent(m.Media(), "", "  ")
		} else if strings.Contains(m.Args(), "-c") {
			jsonString, _ = json.MarshalIndent(m.Channel, "", "  ")
		} else {
			jsonString, _ = json.MarshalIndent(m.OriginalUpdate, "", "  ")
		}
	} else {
		r, err := m.GetReplyMessage()
		if err != nil {
			m.Client.SendMessage("me", "<code>Error:</code> <b>"+err.Error()+"</b>", &telegram.SendOptions{ParseMode: telegram.HTML})
			return nil
		}
		if strings.Contains(m.Args(), "-s") {
			jsonString, _ = json.MarshalIndent(r.Sender, "", "  ")
		} else if strings.Contains(m.Args(), "-m") {
			jsonString, _ = json.MarshalIndent(r.Media(), "", "  ")
		} else if strings.Contains(m.Args(), "-c") {
			jsonString, _ = json.MarshalIndent(r.Channel, "", "  ")
		} else if strings.Contains(m.Args(), "-f") {
			jsonString, _ = json.MarshalIndent(r.File, "", "  ")
		} else {
			jsonString, _ = json.MarshalIndent(r.OriginalUpdate, "", "  ")
		}
	}

	// find all "Data": "<base64>" and decode and replace with actual data
	dataFieldRegex := regexp.MustCompile(`"Data": "([a-zA-Z0-9+/]+={0,2})"`)
	dataFields := dataFieldRegex.FindAllStringSubmatch(string(jsonString), -1)
	for _, v := range dataFields {
		decoded, err := base64.StdEncoding.DecodeString(v[1])
		if err != nil {
			m.Client.SendMessage("me", "<code>Error:</code> <b>"+err.Error()+"</b>", &telegram.SendOptions{ParseMode: telegram.HTML})
			return nil
		}
		jsonString = []byte(strings.ReplaceAll(string(jsonString), v[0], `"Data": "`+string(decoded)+`"`))
	}

	if len(jsonString) > 4095 {
		defer os.Remove("message.json")
		tmpFile, err := os.Create("message.json")
		if err != nil {
			m.Client.SendMessage("me", "<code>Error:</code> <b>"+err.Error()+"</b>", &telegram.SendOptions{ParseMode: telegram.HTML})
			return nil
		}

		_, err = tmpFile.Write(jsonString)
		if err != nil {
			m.Client.SendMessage("me", "<code>Error:</code> <b>"+err.Error()+"</b>", &telegram.SendOptions{ParseMode: telegram.HTML})
			return nil
		}

		_, err = m.Client.SendMedia("me", tmpFile.Name(), &telegram.MediaOptions{Caption: "Message JSON"})
		if err != nil {
			m.Client.SendMessage("me", "<code>Error:</code> <b>"+err.Error()+"</b>", &telegram.SendOptions{ParseMode: telegram.HTML})
		}
	} else {
		m.Client.SendMessage("me", "<pre language='json'>"+string(jsonString)+"</pre>", &telegram.SendOptions{ParseMode: telegram.HTML})
	}

	return nil
}
