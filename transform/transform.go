package transform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	parser "github.com/asyncapi/parser/pkg"
	_ "github.com/asyncapi/parser/pkg/errs"
	"github.com/asyncapi/parser/pkg/models"
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/trigger"
	"github.com/project-flogo/microgateway/api"
)

func TransformToJSON(input, output string) {
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

	data, err := json.MarshalIndent(&model, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	flogo := app.Config{}
	flogo.Name = model.Id
	flogo.Type = "flogo:app"
	flogo.Version = "1.0.0"
	flogo.Description = model.Info.Description
	flogo.AppModel = "1.1.0"

	for i, server := range model.Servers {
		if server.Protocol == "kafka" {
			trig := trigger.Config{
				Id:  fmt.Sprintf("server%d", i),
				Ref: "github.com/project-flogo/contrib/trigger/kafka",
				Settings: map[string]interface{}{
					"brokerUrls": server.Url,
					"trustStore": "",
				},
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
		Name: "default",
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

	data, err = json.MarshalIndent(&flogo, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(output+"/flogo.json", data, 0644)
	if err != nil {
		panic(err)
	}
}
