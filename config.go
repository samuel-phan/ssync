// Copyright 2021 Yahoo Inc.
// Licensed under the terms of the Apache 2.0 License. See LICENSE file in project root for terms.

package main

import (
	"errors"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type UserConf struct {
	Excludes []string
}

type ProjectConf struct {
	Nodes     []string
	Excludes  []string
	Delete    bool
	SudoUser  string   `yaml:"sudo-user"`
	ExtraArgs []string `yaml:"extra-args"`
}

var errProjectDirNotFound = errors.New("project directory not found")

var (
	defaultExcludes = []string{
		".idea",
		".vscode",
		".terraform",
		"*.tfstate.backup",
		"*.py[co]",
		"__pycache__",
	}
)

func loadOrInitUserConf() (*UserConf, error) {
	userHomeDir, _ := os.UserHomeDir()
	userConfFile := path.Join(userHomeDir, ".config/ssync/config.yaml")
	userConfDir := path.Dir(userConfFile)
	err := os.MkdirAll(userConfDir, 0755)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(userConfFile); err == nil {
		// file exists
		return loadUserConf(userConfFile)
	} else if os.IsNotExist(err) {
		// file does *not* exist
		userConf := UserConf{
			Excludes: defaultExcludes,
		}
		err := marshalYamlFile(userConf, userConfFile, 0644)
		if err == nil {
			log.Printf("User configuration file \"%s\" created.", userConfFile)
		}
		return &userConf, err
	} else {
		// file may or may not exist. See err for details.
		return nil, err
	}

}

func loadUserConf(userConfFile string) (*UserConf, error) {
	userConf := UserConf{}
	err := unmarshalYamlFile(&userConf, userConfFile)
	if err != nil {
		return nil, err
	}

	debugf("loadUserConf: %v", userConf)
	return &userConf, nil
}

func getProjectConfFile(confDir string) string {
	return path.Join(confDir, ".ssync")
}

func findProjectDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for dir != "/" {
		debugf("findProjectDir: current dir %s", dir)
		if _, err := os.Stat(getProjectConfFile(dir)); err == nil {
			break
		}
		dir = path.Dir(dir)
	}

	if dir == "/" {
		debugf("findProjectDir: project dir not found.")
		return "", errProjectDirNotFound
	} else {
		debugf("findProjectDir: project dir found: %s", dir)
		return dir, nil
	}
}

func initProjectConf(projectDir string, userConf *UserConf) error {
	projectConfFile := getProjectConfFile(projectDir)
	if _, err := os.Stat(projectConfFile); err == nil {
		// file exists
		log.Printf("ssync configuration file \"%s\" already exists. Nothing to do.", projectConfFile)
		return nil
	} else if os.IsNotExist(err) {
		// file does *not* exist
		projectConf := newDefaultProjectConf(userConf)
		err := writeProjectConfFile(projectConf, projectConfFile)
		if err == nil {
			log.Printf("ssync configuration file \"%s\" created.", projectConfFile)
		}
		return err
	} else {
		// file may or may not exist. See err for details.
		return err
	}
}

func newDefaultProjectConf(userConf *UserConf) *ProjectConf {
	return &ProjectConf{
		Nodes:    []string{"server:/path"},
		Excludes: userConf.Excludes,
		Delete:   true,
	}
}

func loadProjectConf(projectDir string) (*ProjectConf, error) {
	projectConfFile := getProjectConfFile(projectDir)

	projectConf := ProjectConf{}
	err := unmarshalYamlFile(&projectConf, projectConfFile)
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(projectConf.Nodes))
	for _, node := range projectConf.Nodes {
		if node != "" {
			nodes = append(nodes, node)
		}
	}
	projectConf.Nodes = nodes

	debugf("loadProjectConf: %v", projectConf)
	return &projectConf, nil
}

func writeProjectConfFile(projectConf *ProjectConf, confFile string) error {
	return marshalYamlFile(*projectConf, confFile, 0644)
}

func unmarshalYamlFile(obj interface{}, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, obj)
	if err != nil {
		return err
	}

	return nil
}

func marshalYamlFile(obj interface{}, outFile string, mode os.FileMode) error {
	data, err := yaml.Marshal(&obj)
	if err != nil {
		return err
	}
	err = os.WriteFile(outFile, data, mode)
	if err != nil {
		return err
	}
	return nil
}
