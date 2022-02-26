package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	addFirstLabelPatch string = `[
		{ "op": "add", "path": "/metadata/labels", "value": {"added-label": "yes"}}
	]`
	addAdditionalLabelPatch string = `[
         { "op": "add", "path": "/metadata/labels/added-label", "value": "yes" }
     ]`
	updateLabelPatch string = `[
         { "op": "replace", "path": "/metadata/labels/added-label", "value": "yes" }
     ]`
)

type WebhookCfg struct {
	certFile    string
	keyFile     string
	nsBlacklist string
	port        int
	bindAddress string
}

func (whCfg *WebhookCfg) AddFlags(fs *pflag.FlagSet) {
	if whCfg == nil {
		return
	}
	fs.StringVar(&whCfg.certFile, "tls-cert-file", "",
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	fs.StringVar(&whCfg.keyFile, "tls-private-key-file", "",
		"File containing the default x509 private key matching --tls-cert-file.")
	fs.IntVar(&whCfg.port, "port", 443, "Secure port that the webhook listens on.")
	fs.StringVar(&whCfg.nsBlacklist, "exclude-namespaces", "", "Comma separated namespace blacklist.")
	fs.StringVar(&whCfg.bindAddress, "bind-address", "0.0.0.0", "The IP address on which to listen.")

}

func toV1AdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func changeRuntimeClass(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	klog.V(2).Info("calling add-label")
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}

	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	pt := v1beta1.PatchTypeJSONPatch
	labelValue, hasLabel := obj.ObjectMeta.Labels["added-label"]
	switch {
	case len(obj.ObjectMeta.Labels) == 0:
		reviewResponse.Patch = []byte(addFirstLabelPatch)
		reviewResponse.PatchType = &pt
	case !hasLabel:
		reviewResponse.Patch = []byte(addAdditionalLabelPatch)
		reviewResponse.PatchType = &pt
	case labelValue != "yes":
		reviewResponse.Patch = []byte(updateLabelPatch)
		reviewResponse.PatchType = &pt
	default:
		// already set
	}
	return &reviewResponse
}

func changeRuntimeClassV1(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("calling add-label")
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}

	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	pt := v1.PatchTypeJSONPatch
	labelValue, hasLabel := obj.ObjectMeta.Labels["added-label"]
	switch {
	case len(obj.ObjectMeta.Labels) == 0:
		reviewResponse.Patch = []byte(addFirstLabelPatch)
		reviewResponse.PatchType = &pt
	case !hasLabel:
		reviewResponse.Patch = []byte(addAdditionalLabelPatch)
		reviewResponse.PatchType = &pt
	case labelValue != "yes":
		reviewResponse.Patch = []byte(updateLabelPatch)
		reviewResponse.PatchType = &pt
	default:
		// already set
	}
	return &reviewResponse
}

func changeRuntimeClassHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(2).Infof(fmt.Sprintf("handling request: %s", body))

	deserializer := codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1beta1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissonReview := &v1beta1.AdmissionReview{}
		responseAdmissonReview.SetGroupVersionKind(*gvk)
		responseAdmissonReview.Response = changeRuntimeClass(requestedAdmissionReview)
		responseAdmissonReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissonReview
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			klog.Error("Expected v1.AdmmisionReview but got %T", obj)
			return
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = changeRuntimeClassV1(requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseObj))
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func (whCfg WebhookCfg) WebHookServer() {

	go func() {
		http.HandleFunc("/mutate", changeRuntimeClassHandler)
		err := http.ListenAndServeTLS("0.0.0.0:443", whCfg.certFile, whCfg.keyFile, nil)
		if err != nil {
			klog.Error(err)
		}
	}()
	klog.V(0).Info("webhook server started")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	fmt.Printf("Get signal %s, exit ...", s.String())
	close(quit)
}
