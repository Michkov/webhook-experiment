package admissionwebhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"k8s.io/api/admission/v1beta1"
)

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	// read the body / request
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		sendError(err, w)
		return
	}

	// mutate the request
	verbose := true
	if verbose {
		log.Printf("recv: %s\n", string(body)) // untested section
	}

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		sendError(fmt.Errorf("unmarshaling request failed with %s", err), w)
	}
	ar := admReview.Request

	responseBody := []byte{}

	if ar.Kind.Kind == "PipelineRun" && ar.Kind.Group == "tekton.dev" {
		responseBody, err = ProcessPipelineRun(admReview, verbose)
	}

	if err != nil {
		sendError(err, w)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}
