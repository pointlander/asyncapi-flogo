module github.com/pointlander/asyncapi-flogo

go 1.12

replace (
	github.com/project-flogo/core => github.com/pointlander/core v0.9.0-alpha.0.0.20190521204626-d0604d8121c1
	github.com/project-flogo/edge-contrib/trigger/mqtt => github.com/pointlander/edge-contrib/trigger/mqtt v0.0.0-20190523190809-4fd354c541c2
)

require (
	github.com/asyncapi/parser v0.0.0-20190506150237-e2e785dfad03
	github.com/nareshkumarthota/flogocomponents v0.0.0-20190410061230-d24c4239918a
	github.com/project-flogo/cli v0.9.0-rc.2
	github.com/project-flogo/contrib/activity/kafka v0.9.1-0.20190516180541-534215f1b7ac
	github.com/project-flogo/contrib/activity/log v0.9.0-rc.1.0.20190509204259-4246269fb68e
	github.com/project-flogo/contrib/activity/rest v0.9.0-rc.1.0.20190509204259-4246269fb68e
	github.com/project-flogo/contrib/trigger/kafka v0.9.1-0.20190603184501-d845e1d612f8
	github.com/project-flogo/contrib/trigger/rest v0.9.0-rc.1.0.20190509204259-4246269fb68e
	github.com/project-flogo/core v0.9.0
	github.com/project-flogo/edge-contrib/activity/mqtt v0.0.0-20190521185544-b79879165f97
	github.com/project-flogo/edge-contrib/trigger/mqtt v0.0.0-20190521185544-b79879165f97
	github.com/project-flogo/eftl v0.0.0-20190318194200-d6dc627012e5
	github.com/project-flogo/microgateway v0.0.0-20190521205136-e8ff8943422a
	github.com/project-flogo/websocket v0.0.0-20190201184711-2efafcb15730
	github.com/spf13/cobra v0.0.3
)
