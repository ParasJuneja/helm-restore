package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	utils "github.com/ParasJuneja/helm-plugin-utils/helm-utils"
	"gopkg.in/yaml.v2"
)

type Y map[interface{}]interface{}

func Restore(releaseName string) error {
	err, releaseInfo := utils.GetRelease(releaseName)
	if err != nil {
		return err
	}
	releaseNamespace := releaseInfo.Namespace
	resources := strings.Split(releaseInfo.Manifest, "---")
	for _, resource := range resources {
		data := make(Y)
		if err := yaml.Unmarshal([]byte(resource), &data); err != nil {
			return err
		}
		if resource != "" {
			kind := data["kind"].(string)
			name := data["metadata"].(Y)["name"]
			namespace := data["metadata"].(Y)["namespace"]
			operation := "replace"
			if name != nil {
				name = name.(string)
			} else {
				name = releaseName
			}
			if namespace != nil {
				namespace = namespace.(string)
			} else {
				namespace = releaseNamespace
			}
			//fmt.Println(resource)
			if kind == "Service" {
				getClusterIP := exec.Command("kubectl", "get", "service", name.(string), "-n", namespace.(string), "-o", "jsonpath='{.spec.clusterIP}'")
				clusterIP, err := getClusterIP.Output()
				if err != nil {
					return err
				}
				data["spec"].(Y)["clusterIP"] = strings.ReplaceAll(string(clusterIP), "'", "")
			} else if data["kind"] == "PersistentVolumeClaim" {
				operation = "apply"
			}
			d, err := yaml.Marshal(data)
			if err != nil {
				return err
			}
			resource = strings.ReplaceAll(string(d), "$", "\\$")
			if err := ioutil.WriteFile("/tmp/manifest.yaml", []byte(resource), 0666); err != nil {
				return err
			}

			commandToExecute := exec.Command("kubectl", operation, "-f", "/tmp/manifest.yaml", "-n", namespace.(string))
			commandToExecute.Stdout = os.Stdout
			commandToExecute.Stderr = os.Stderr
			if err := commandToExecute.Run(); err != nil {
				return err
			}
		}
	}
	if err := os.Remove("/tmp/manifest.yaml"); err != nil {
		return err
	}
	return nil
}
