package main

var config = Config{
	StudentConfigFilename: "config.yml",
	LmsBaseUrl:            "https://lms.metaclass.kts.studio/api/v2.chunk.check_callback_task",

	GithubRepo:        "${ORIGINAL_REPOSITORY}",
	LmsCompanyToken:   "${LMS_COMPANY_TOKEN}",
	CallbackTaskId:    "${CALLBACK_TASK_ID}",
	GithubStudentRepo: "${GITHUB_REPOSITORY}",
	GithubStudentRef:  "${GITHUB_REF}",
}
