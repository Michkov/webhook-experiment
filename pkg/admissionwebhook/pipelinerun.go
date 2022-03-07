package admissionwebhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tektonapi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProcessPipelineRun(admReview v1beta1.AdmissionReview, verbose bool) ([]byte, error) {
	ar := admReview.Request
	var err error
	var pipelineRun *tektonapi.PipelineRun

	resp := v1beta1.AdmissionResponse{}

	// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
	if err := json.Unmarshal(ar.Object.Raw, &pipelineRun); err != nil {
		return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
	}
	// set response options
	resp.Allowed = true
	resp.UID = ar.UID
	pT := v1beta1.PatchTypeJSONPatch
	resp.PatchType = &pT // it's annoying that this needs to be a pointer as you cannot give a pointer to a constant?

	p := []map[string]string{}
	bundle := pipelineRun.Spec.PipelineRef.Bundle
	if bundle != "" {
		log.Printf("Bundle %s defined", bundle)
	} else {
		client, _ := createClient("")
		configMap, _ := client.CoreV1().ConfigMaps(ar.Namespace).Get(context.TODO(), "build-pipelines-defaults", metav1.GetOptions{})
		detected_bundle := configMap.Data["default_build_bundle"]
		if detected_bundle == "" {
			configMap, _ = client.CoreV1().ConfigMaps("build-templates").Get(context.TODO(), "build-pipelines-defaults", metav1.GetOptions{})
			detected_bundle = configMap.Data["default_build_bundle"]
		}
		fmt.Println(detected_bundle)
		patch := map[string]string{
			"op":    "add",
			"path":  "/spec/pipelineRef/bundle",
			"value": detected_bundle,
		}
		p = append(p, patch)
	}
	// parse the []map into JSON
	resp.Patch, _ = json.Marshal(p)

	// Success, of course ;)
	resp.Result = &metav1.Status{
		Status: "Success",
	}

	admReview.Response = &resp
	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	responseBody, err := json.Marshal(admReview)
	if err != nil {
		return nil, err // untested section
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody)) // untested section
	}
	return responseBody, nil
}
