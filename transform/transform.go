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

	for i, server := range model.Servers {
		if isSecure := server.Protocol == "kafka-secure"; server.Protocol == "kafka" || isSecure {
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
			trig := trigger.Config{
				Id:  fmt.Sprintf("server%d", i),
				Ref: "github.com/project-flogo/contrib/trigger/kafka",
				Settings: map[string]interface{}{
					"brokerUrls": server.Url,
					"trustStore": "",
				},
			}
			if hasUserPassword {
				trig.Settings["user"] = "=$env[USER]"
				trig.Settings["password"] = "=$env[PASSWORD]"
			}
			if isSecure {
				trig.Settings["trustStore"] = "=$env[TRUST_STORE]"
			}
			for name, channel := range model.Channels {
				if channel.Subscribe != nil {
					handler := trigger.HandlerConfig{
						Settings: map[string]interface{}{
							"topic": name,
						},
					}
					action := action.Config{
						Ref: "github.com/project-flogo/microgateway",
						Settings: map[string]interface{}{
							"uri": "microgateway:Default",
						},
					}
					actionConfig := trigger.ActionConfig{
						Config: &action,
					}
					handler.Actions = append(handler.Actions, &actionConfig)
					trig.Handlers = append(trig.Handlers, &handler)
				}
			}
			flogo.Triggers = append(flogo.Triggers, &trig)
		}
	}

	gateway := api.Microgateway{
		Name: "Default",
	}
	service := api.Service{
		Name:        "log",
		Ref:         "github.com/project-flogo/contrib/activity/log",
		Description: "logging service",
	}
	gateway.Services = append(gateway.Services, &service)
	step := api.Step{
		Service: "log",
		Input: map[string]interface{}{
			"message": "test message",
		},
	}
	gateway.Steps = append(gateway.Steps, &step)

	raw, err := json.Marshal(&gateway)
	if err != nil {
		panic(err)
	}

	res := resource.Config{
		ID:   "microgateway:Default",
		Data: raw,
	}
	flogo.Resources = append(flogo.Resources, &res)

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
