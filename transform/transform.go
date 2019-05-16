package transform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	parser "github.com/asyncapi/parser/pkg"
	_ "github.com/asyncapi/parser/pkg/errs"
	"github.com/asyncapi/parser/pkg/models"
	"github.com/project-flogo/core/action"
	coreapi "github.com/project-flogo/core/api"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/trigger"
	_ "github.com/project-flogo/microgateway"
	"github.com/project-flogo/microgateway/api"
)

// Transform converts an asyn api to a new representation
func Transform(input, output, conversionType string) {
	switch conversionType {
	case "flogoapiapp":
		ToAPI(input, output)
	case "flogodescriptor":
		ToJSON(input, output)
	default:
		panic("invalid type")
	}
}

type protocolConfig struct {
	name, secure      string
	trigger, activity string
	port              string
	urlSetting        string
	trustStoreSetting string
	topicSetting      string
	contentPath       string
}

func (p protocolConfig) protocol(model *models.AsyncapiDocument, schemes map[string]interface{}, flogo *app.Config) {
	services := make([]*api.Service, 0, 8)
	for i, server := range model.Servers {
		if isSecure := server.Protocol == p.secure; server.Protocol == p.name || isSecure {
			hasUserPassword := false
			for _, requirement := range server.Security {
				for scheme := range *requirement {
					if entry := schemes[scheme]; entry != nil {
						if definition, ok := entry.(map[string]interface{}); ok {
							if value := definition["type"]; value != nil {
								if typ, ok := value.(string); ok && typ == "userPassword" {
									hasUserPassword = true
								}
							}
						}
					}
				}
			}
			brokerUrls := fmt.Sprintf("%s%dURL", p.name, i)
			attribute := data.NewAttribute(brokerUrls, data.TypeString, server.Url)
			brokerUrls = fmt.Sprintf("=$property[%s]", brokerUrls)
			flogo.Properties = append(flogo.Properties, attribute)
			trig := trigger.Config{
				Id:  fmt.Sprintf("%s%d", p.name, i),
				Ref: p.trigger,
				Settings: map[string]interface{}{
					p.urlSetting: brokerUrls,
				},
			}
			if p.name == "eftl" || p.name == "eftl-secure" {
				trig.Settings["id"] = p.name
			}
			if hasUserPassword {
				trig.Settings["user"] = "=$env[USER]"
				trig.Settings["password"] = "=$env[PASSWORD]"
			}
			if isSecure {
				trig.Settings[p.trustStoreSetting] = "=$env[TRUST_STORE]"
			}
			for name, channel := range model.Channels {
				if channel.Subscribe != nil {
					settings := map[string]interface{}{
						p.topicSetting: name,
					}
					if len(channel.Subscribe.ProtocolInfo) > 0 {
						var protocolInfo map[string]interface{}
						err := json.Unmarshal(channel.Subscribe.ProtocolInfo, &protocolInfo)
						if err != nil {
							panic(err)
						}
						if value := protocolInfo["flogo-kafka"]; value != nil {
							if flogo, ok := value.(map[string]interface{}); ok {
								if value := flogo["partitions"]; value != nil {
									if partitions, ok := value.(string); ok {
										settings["partitions"] = partitions
									}
								}
								if value := flogo["offset"]; value != nil {
									if offset, ok := value.(float64); ok {
										settings["offset"] = int64(offset)
									}
								}
							}
						}
					}
					handler := trigger.HandlerConfig{
						Settings: settings,
					}
					action := action.Config{
						Ref: "github.com/project-flogo/microgateway",
						Settings: map[string]interface{}{
							"uri": fmt.Sprintf("microgateway:%s", p.name),
						},
					}
					actionConfig := trigger.ActionConfig{
						Config: &action,
					}
					handler.Actions = append(handler.Actions, &actionConfig)
					trig.Handlers = append(trig.Handlers, &handler)
				}
				if channel.Publish != nil {
					settings := map[string]interface{}{
						p.urlSetting:   brokerUrls,
						p.topicSetting: name,
					}
					if p.name == "eftl" || p.name == "eftl-secure" {
						settings["id"] = p.name
					}
					if hasUserPassword {
						settings["user"] = "=$.env[USER]"
						settings["password"] = "=$.env[PASSWORD]"
					}
					if isSecure {
						settings[p.trustStoreSetting] = "=$.env[TRUST_STORE]"
					}
					service := &api.Service{
						Name:        fmt.Sprintf("%s-name-%s", p.name, name),
						Ref:         p.activity,
						Description: fmt.Sprintf("%s service", p.name),
						Settings:    settings,
					}
					services = append(services, service)
				}
			}
			flogo.Triggers = append(flogo.Triggers, &trig)
		}
	}

	if len(flogo.Triggers) > 0 {
		gateway := &api.Microgateway{
			Name: p.name,
		}
		service := &api.Service{
			Name:        "log",
			Ref:         "github.com/project-flogo/contrib/activity/log",
			Description: "logging service",
		}
		gateway.Services = append(gateway.Services, service)
		step := &api.Step{
			Service: "log",
			Input: map[string]interface{}{
				"message": fmt.Sprintf("=$.payload.%s", p.contentPath),
			},
		}
		gateway.Steps = append(gateway.Steps, step)

		raw, err := json.Marshal(gateway)
		if err != nil {
			panic(err)
		}

		res := &resource.Config{
			ID:   fmt.Sprintf("microgateway:%s", p.name),
			Data: raw,
		}
		flogo.Resources = append(flogo.Resources, res)
	}

	if len(services) > 0 {
		trig := trigger.Config{
			Id:  fmt.Sprintf("%sPublish", p.name),
			Ref: "github.com/project-flogo/contrib/trigger/rest",
			Settings: map[string]interface{}{
				"port": p.port,
			},
		}
		handler := trigger.HandlerConfig{
			Settings: map[string]interface{}{
				"method": "POST",
				"path":   "/post",
			},
		}
		action := action.Config{
			Ref: "github.com/project-flogo/microgateway",
			Settings: map[string]interface{}{
				"uri": fmt.Sprintf("microgateway:%sPublish", p.name),
			},
		}
		actionConfig := trigger.ActionConfig{
			Config: &action,
		}
		handler.Actions = append(handler.Actions, &actionConfig)
		trig.Handlers = append(trig.Handlers, &handler)
		flogo.Triggers = append(flogo.Triggers, &trig)

		gateway := &api.Microgateway{
			Name: fmt.Sprintf("%sPublish", p.name),
		}
		service := &api.Service{
			Name:        "log",
			Ref:         "github.com/project-flogo/contrib/activity/log",
			Description: "logging service",
		}
		gateway.Services = append(services, service)
		step := &api.Step{
			Service: "log",
			Input: map[string]interface{}{
				"message": "=$.payload.content",
			},
		}
		gateway.Steps = append(gateway.Steps, step)

		raw, err := json.Marshal(gateway)
		if err != nil {
			panic(err)
		}

		res := &resource.Config{
			ID:   fmt.Sprintf("microgateway:%sPublish", p.name),
			Data: raw,
		}
		flogo.Resources = append(flogo.Resources, res)
	}
}

func convert(input string) *app.Config {
	document, err := ioutil.ReadFile(input)
	if err != nil {
		panic(err)
	}

	parsed, perr := parser.Parse(document, true)
	if perr != nil {
		panic(err)
	}

	model := models.AsyncapiDocument{}
	err = json.Unmarshal(parsed, &model)
	if err != nil {
		panic(err)
	}

	flogo := app.Config{}
	flogo.Name = model.Id
	flogo.Type = "flogo:app"
	flogo.Version = "1.0.0"
	flogo.Description = model.Info.Description
	flogo.AppModel = "1.1.0"

	var schemes map[string]interface{}
	if len(model.Components.SecuritySchemes) > 0 {
		err = json.Unmarshal(model.Components.SecuritySchemes, &schemes)
		if err != nil {
			panic(err)
		}
	}

	configs := [...]protocolConfig{
		{
			name:              "kafka",
			secure:            "kafka-secure",
			trigger:           "github.com/project-flogo/contrib/trigger/kafka",
			activity:          "github.com/project-flogo/contrib/activity/kafka",
			port:              "9096",
			urlSetting:        "brokerUrls",
			trustStoreSetting: "trustStore",
			topicSetting:      "topic",
			contentPath:       "message",
		},
		{
			name:              "eftl",
			secure:            "eftl-secure",
			trigger:           "github.com/project-flogo/eftl/trigger",
			activity:          "github.com/project-flogo/eftl/activity",
			port:              "9097",
			urlSetting:        "url",
			trustStoreSetting: "ca",
			topicSetting:      "dest",
			contentPath:       "content",
		},
	}
	for _, config := range configs {
		config.protocol(&model, schemes, &flogo)
	}

	return &flogo
}

// ToAPI converts an asyn api to a API flogo application
func ToAPI(input, output string) {
	flogo := convert(input)
	coreapi.Generate(flogo, output+"/app.go")
}

// ToJSON converts an async api to a JSON flogo application
func ToJSON(input, output string) {
	flogo := convert(input)
	data, err := json.MarshalIndent(flogo, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(output+"/flogo.json", data, 0644)
	if err != nil {
		panic(err)
	}
}
