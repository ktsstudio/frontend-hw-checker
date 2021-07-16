package main

var config = Config{
	TestsDirectoryName:   "tests",
	PipelineFilename:     ".github/workflows/pipeline.yml",
	SkillsConfigFilename: "skills_config.yml",
	SkillsBaseUrl:        "https://skills.kube1.ktsdev.ru/api/v2.chunk.check_callback_task",

	ZipHash:      "${ZIP_HASH}",
	PipelineHash: "${PIPELINE_HASH}",

	SkillsAuthToken: "${SKILLS_AUTH_TOKEN}",
	CallbackTaskId:  "${CALLBACK_TASK_ID}",
}
