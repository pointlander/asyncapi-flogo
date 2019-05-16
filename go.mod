module github.com/pointlander/asyncapi-flogo

go 1.12

replace (
	github.com/project-flogo/contrib/activity/kafka => github.com/pointlander/contrib/activity/kafka v0.0.0-20190516172203-358320c7771c
	github.com/project-flogo/core => github.com/pointlander/core v0.9.0-alpha.0.0.20190516170615-e0b906d347b3
)

require (
	github.com/asyncapi/parser v0.0.0-20190506150237-e2e785dfad03
	github.com/project-flogo/cli v0.9.0-rc.2
	github.com/project-flogo/contrib/activity/kafka v0.9.0
	github.com/project-flogo/contrib/activity/log v0.9.0-rc.1.0.20190509204259-4246269fb68e
	github.com/project-flogo/contrib/trigger/kafka v0.9.1
	github.com/project-flogo/contrib/trigger/rest v0.9.0-rc.1.0.20190509204259-4246269fb68e
	github.com/project-flogo/core v0.9.0
	github.com/project-flogo/microgateway v0.0.0-20190514214306-204c38dcda09
	github.com/spf13/cobra v0.0.3
)
