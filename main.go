package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type Config struct {
	StudentConfigFilename string

	GithubRepo      string
	LmsCompanyToken string
	LmsBaseUrl      string
	CallbackTaskId  string

	GithubStudentRepo string
	GithubStudentRef  string
}

var PYTEST_RESULT_PATTERN = regexp.MustCompile("={25}\\s*(?P<failed>\\d+ failed,?)?\\s*(?P<passed>\\d+ passed,?)?\\s*(?P<skipped>\\d+ skipped,?)? in .*={25}")

type StudentConfig struct {
	UserToken string `yaml:"user_token"`
}

func getStudentConfig() (StudentConfig, error) {
	studentConfig := StudentConfig{}
	f, err := os.Open(config.StudentConfigFilename)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		log.Printf("Can not open %s", config.StudentConfigFilename)
		return StudentConfig{}, err
	}

	log.Printf("Reading %s", config.StudentConfigFilename)
	buffer, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Can not read %s", config.StudentConfigFilename)
		return StudentConfig{}, err
	}

	err = yaml.Unmarshal(buffer, &studentConfig)
	log.Printf("Parse %s", config.StudentConfigFilename)
	if err != nil {
		log.Printf("Can not parse %s", config.StudentConfigFilename)
		return StudentConfig{}, err
	}
	return studentConfig, nil
}

func submitResult(skillsConfig StudentConfig) error {
	client := http.Client{}

	type RequestExtra struct {
		StudentRepo string `json:"student_repo"`
		StudentRef  string `json:"student_ref"`
	}

	type Request struct {
		Created   time.Time    `json:"created"`
		TaskID    string       `json:"task_id"`
		UserToken string       `json:"user_token"`
		Extra     RequestExtra `json:"extra"`
	}
	data, err := json.Marshal(Request{
		Created:   time.Now(),
		TaskID:    config.CallbackTaskId,
		UserToken: skillsConfig.UserToken,
		Extra: RequestExtra{
			StudentRepo: config.GithubStudentRepo,
			StudentRef:  config.GithubStudentRef,
		},
	})
	req, err := http.NewRequest("POST", config.LmsBaseUrl, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.LmsCompanyToken))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Response status code %d instead of 200", resp.StatusCode))
	}
	return nil
}

func runCommand(cmd string, args ...string) (bytes.Buffer, error) {
	var outBuffer bytes.Buffer
	execCmd := exec.Command(cmd, args...)
	execCmd.Stdout = &outBuffer
	execCmd.Stderr = &outBuffer

	err := execCmd.Run()
	if outBuffer.Len() > 0 {
		log.Println(outBuffer.String())
	}
	if err != nil {
		return bytes.Buffer{}, err
	}
	return outBuffer, nil
}

func setupEnvironment() error {
	var err error
	log.Printf("Setuping tests environment")
	if _, err = os.Stat("requirements.txt"); err == nil {
		log.Printf("Found requirements.txt")
		log.Printf("Run pip3 install -r requirements.txt")
		_, err = runCommand("pip3", "install", "-r", "requirements.txt")
		if err != nil {
			log.Printf("Can not install requirements: %e", err)
			return err
		}
	}

	log.Printf("Installing pytest")
	_, err = runCommand("pip3", "install", "pytest")
	if err != nil {
		log.Printf("Can not install pytest: %e", err)
		return err
	}
	log.Printf("Removing student tests/ dir")
	_, err = runCommand("rm", "-rf", "tests")
	if err != nil {
		log.Printf("Can not remove tests/ folder: %e", err)
		return err
	}

	log.Printf("Cloning original repo")
	_, err = runCommand("git", "clone", config.GithubRepo, "original_repo")
	if err != nil {
		log.Printf("Can not clone original repo: %e", err)
		return err
	}

	log.Printf("Moving tests folder to student code")
	_, err = runCommand("mv", "original_repo/tests", "tests/")
	if err != nil {
		log.Printf("Can not move original tests/ folder to student code")
		return err
	}
	return nil
}

func validatePytestOutput(output string) error {
	match := PYTEST_RESULT_PATTERN.FindStringSubmatch(output)
	result := make(map[string]string)
	for i, name := range PYTEST_RESULT_PATTERN.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	if result["failed"] != "" || result["skipped"] != "" {
		fmt.Printf("Not all tests are passed")
		return errors.New("some tests are failed or skipped")
	}
	return nil
}

func runPytest() error {
	buffer, err := runCommand("pytest", "tests/")
	if err != nil {
		log.Printf("Can not run pytest: %e", err)
		return err
	}
	err = validatePytestOutput(buffer.String())
	if err != nil {
		log.Printf("Can not validate tests result: %e", err)
		return err
	}
	return nil
}

func main() {
	log.Printf("Searching for %s", config.StudentConfigFilename)
	skillsConfig, err := getStudentConfig()
	if err != nil {
		log.Fatalf("Can not read %s: %v", config.StudentConfigFilename, err)
	}
	if skillsConfig.UserToken == "" {
		log.Fatalf(fmt.Sprintf("Can not find user_token in %s", config.StudentConfigFilename))
	}

	err = setupEnvironment()
	if err != nil {
		log.Fatal("Can not setup environment")
	}

	err = runPytest()
	if err != nil {
		log.Fatal("Can not run pytest")
	}

	log.Printf("Submiting success result")
	err = submitResult(skillsConfig)
	if err != nil {
		log.Fatalf("Can not submit result: %e. Please try again later.", err)
	}
}
