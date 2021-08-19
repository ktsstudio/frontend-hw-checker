package main

import (
	"validator/v1/pkg/app_config"
)

var config = app_config.Config{
	LmsBaseUrl: "https://lms.metaclass.kts.studio/api/v2.chunk.check_callback_task",

	GithubRepo:      "${ORIGINAL_REPOSITORY}",
	LmsCompanyToken: "${LMS_COMPANY_TOKEN}",
	CallbackTaskId:  "${CALLBACK_TASK_ID}",
	StudentConfig: app_config.StudentConfig{
		ConfigFilename: "config.yml",
		UserToken:      "",
		StudentRepo:    "",
		StudentRef:     "",
	},
	S3Config: app_config.S3Config{
		Region:          "ru-msk",
		EndpointUrl:     "https://hb.bizmrg.com/",
		AccessKeyID:     "hjeY7WHbD3fSG7iNJxwpEu",
		SecretAccessKey: "fdqucBuuGP9WDqXjspnV7weUjNWxtT1Tf7ptgZwWjHtH",
		Bucket:          "lms-metaclass-prod-student-data",
	},
}
