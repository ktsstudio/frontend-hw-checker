package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/mod/sumdb/dirhash"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	TestsDirectoryName   string
	PipelineFilename     string
	SkillsConfigFilename string

	ZipHash      string
	PipelineHash string

	SkillsAuthToken string
	SkillsBaseUrl   string
	CallbackTaskId  string
}

type SkillsConfig struct {
	UserToken string `yaml:"user_token"`
}

const (
	ZipFilename = "tests.zip"
)

func zipDirectory(folder, zipName string) error {
	destinationFile, err := os.Create(zipName)
	if err != nil {
		return err
	}
	zipArchive := zip.NewWriter(destinationFile)
	err = filepath.Walk(folder, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(folder))
		zipFile, err := zipArchive.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		_ = fsFile.Close()
		return nil
	})
	if err != nil {
		return err
	}
	err = zipArchive.Close()
	if err != nil {
		return err
	}
	return nil
}

func hashFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	reader := io.Reader(f)

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum), nil
}

func getSkillsConfig() (SkillsConfig, error) {
	skillsConfig := SkillsConfig{}
	f, err := os.Open(config.SkillsConfigFilename)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return SkillsConfig{}, err
	}

	buffer, err := ioutil.ReadAll(f)
	if err != nil {
		return SkillsConfig{}, err
	}

	err = yaml.Unmarshal(buffer, &skillsConfig)
	return skillsConfig, err
}

func submitResult(skillsConfig SkillsConfig) error {
	client := http.Client{}

	type Request struct {
		Created   time.Time `json:"created"`
		TaskID    string    `json:"task_id"`
		UserToken string    `json:"user_token"`
	}
	data, err := json.Marshal(Request{
		Created:   time.Now(),
		TaskID:    config.CallbackTaskId,
		UserToken: skillsConfig.UserToken,
	})
	req, err := http.NewRequest("POST", config.SkillsBaseUrl, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.SkillsAuthToken))
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

func main() {
	skillsConfig, err := getSkillsConfig()
	if err != nil {
		log.Fatalf("Can not parse %s: %v", config.SkillsConfigFilename, err)
	}
	if skillsConfig.UserToken == "" {
		log.Fatalf(fmt.Sprintf("Can not find user_token in %s", config.SkillsConfigFilename))
	}

	if _, err = os.Stat(config.TestsDirectoryName); os.IsNotExist(err) {
		log.Fatalf(fmt.Sprintf("Directory '%s' does not exist", config.TestsDirectoryName))
	}

	err = zipDirectory(config.TestsDirectoryName, ZipFilename)
	if err != nil {
		log.Fatalf("Can not zip the '%s' directory: %v", config.TestsDirectoryName, err)
	}

	zipHash, err := dirhash.HashZip(ZipFilename, dirhash.Hash1)
	if err != nil {
		log.Fatalf("Can not get hash of the '%s': %v", config.TestsDirectoryName, err)
	}

	if zipHash != config.ZipHash {
		log.Fatalf("Directory %s has been modified. Please restore it to original state.", config.TestsDirectoryName)
	}

	pipelineHash, err := hashFile(config.PipelineFilename)
	if err != nil {
		log.Fatalf("Can not get hash of the 'pipeline.yml': %v", err)
	}

	if pipelineHash != config.PipelineHash {
		log.Fatalf("File %s has been modified. Please restore it to original state.", config.PipelineFilename)
	}

	err = submitResult(skillsConfig)
	if err != nil {
		log.Fatalf("Can not submit results to Skills. Please try again later.")
	}
}
