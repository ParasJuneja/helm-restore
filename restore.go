package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	utils "github.com/ParasJuneja/helm-plugin-utils/helm-utils"
	"gopkg.in/yaml.v2"
)

func Restore(releaseName string) error {
	err, releaseInfo := utils.GetRelease(releaseName)
	if err != nil {
		return err
	}
	releaseNamespace := releaseInfo.Namespace
	resources := strings.Split(releaseInfo.Manifest, "---")
	data := make(map[interface{}]interface{})
	for _, resource := range resources {
		if err := yaml.Unmarshal([]byte(resource), &data); err != nil {
			return err
		}
		if resource != "" {
			name := data["metadata"].(map[interface{}]interface{})["name"]
			namespace := data["metadata"].(map[interface{}]interface{})["namespace"]
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
			if data["kind"] == "Service" {
				getClusterIP := exec.Command("kubectl", "get", "service", name.(string), "-n", namespace.(string), "-o", "jsonpath='{.spec.clusterIP}'")
				clusterIP, err := getClusterIP.Output()
				if err != nil {
					return err
				}
				data["spec"].(map[interface{}]interface{})["clusterIP"] = strings.ReplaceAll(string(clusterIP), "'", "")
			}
			d, err := yaml.Marshal(data)
			if err != nil {
				return err
			}
			resource = strings.ReplaceAll(string(d), "$", "\\$")
			if err := ioutil.WriteFile("/tmp/manifest.yaml", []byte(resource), 0666); err != nil {
				return err
			}

			commandToExecute := exec.Command("kubectl", "replace", "-f", "/tmp/manifest.yaml", "-n", namespace.(string))
			commandToExecute.Stdout = os.Stdout
			commandToExecute.Stderr = os.Stderr
			if err := commandToExecute.Run(); err != nil {
				return err
			}

		}
	}
	os.Remove("/tmp/manifest.yaml")
	return nil
}
