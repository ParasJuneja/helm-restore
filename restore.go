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

func addClusterIP(name string, namespace string, data Y) error {
	getClusterIP := exec.Command("kubectl", "get", "service", name, "-n", namespace, "-o", "jsonpath='{.spec.clusterIP}'")
	clusterIP, err := getClusterIP.Output()
	if err != nil {
		return err
	}
	data["spec"].(Y)["clusterIP"] = strings.ReplaceAll(string(clusterIP), "'", "")

	return nil
}

func restore(operation string, namespace string) error {
	commandToExecute := exec.Command("kubectl", operation, "-f", "/tmp/manifest.yaml", "-n", namespace)
	commandToExecute.Stdout = os.Stdout
	commandToExecute.Stderr = os.Stderr
	if err := commandToExecute.Run(); err != nil {
		return err
	}
	return nil
}

func delete_empty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func addHelmMeta(releaseName string, releaseNamespace string, data Y) {
	labels := data["metadata"].(Y)["labels"]
	annotations := data["metadata"].(Y)["annotations"]
	if labels == nil {
		data["metadata"].(Y)["labels"] = make(Y)
	}
	if annotations == nil {
		data["metadata"].(Y)["annotations"] = make(Y)
	}
	data["metadata"].(Y)["labels"].(Y)["app.kubernetes.io/managed-by"] = "Helm"
	data["metadata"].(Y)["annotations"].(Y)["meta.helm.sh/release-name"] = releaseName
	data["metadata"].(Y)["annotations"].(Y)["meta.helm.sh/release-namespace"] = releaseNamespace
}

func Restore(releaseName string) error {
	// Get Release info
	err, releaseInfo := utils.GetRelease(releaseName)
	if err != nil {
		return err
	}

	// Extract release namespace
	releaseNamespace := releaseInfo.Namespace

	// Split manifests
	resources := strings.Split(releaseInfo.Manifest, "---")
	//remove empty elements in slice
	resources = delete_empty(resources)

	for _, resource := range resources {
		// read manifest yml
		data := make(Y)
		if err := yaml.Unmarshal([]byte(resource), &data); err != nil {
			return err
		}

		// extract kind, name and namespace
		kind := data["kind"].(string)
		name := data["metadata"].(Y)["name"]
		namespace := data["metadata"].(Y)["namespace"]

		// default restore method
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

		// add immutable clusterIP field for services
		if kind == "Service" {
			if err := addClusterIP(name.(string), namespace.(string), data); err != nil {
				return err
			}
		} else if data["kind"] == "PersistentVolumeClaim" {
			// because PVC spec is immutablem except for resource requests
			operation = "apply"
		}

		// Add Helm Labels and Annotations
		addHelmMeta(releaseName, releaseNamespace, data)

		// map to yml
		d, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		resource = strings.ReplaceAll(string(d), "$", "\\$")

		// Write resource manifest to temporary file
		if err := ioutil.WriteFile("/tmp/manifest.yaml", []byte(resource), 0666); err != nil {
			return err
		}

		//restore resource
		if err := restore(operation, namespace.(string)); err != nil {
			return err
		}
	}
	if err := os.Remove("/tmp/manifest.yaml"); err != nil {
		return err
	}
	return nil
}
