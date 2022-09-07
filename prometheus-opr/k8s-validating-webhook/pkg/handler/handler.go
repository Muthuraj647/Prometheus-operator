package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	admission "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var Clientset kubernetes.Interface

var Env_name string
var Env_value string
var SA_Annotation string

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	Clientset = clientset
}

func Validation(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Println("Method Not Allowed")

		RecordRequests("GET", "401")

		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}
	klog.Info("/Validate")
	ar := new(admission.AdmissionReview)
	err := json.NewDecoder(r.Body).Decode(ar)
	if err != nil {
		handleErr(w, nil, err)
		return
	}

	response := &admission.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: true,
	}

	pod := &corev1.Pod{}

	if err := json.Unmarshal(ar.Request.Object.Raw, pod); err != nil {
		handleErr(w, ar, err)
		return
	}

	RecordRequests("GET", "200")

	//example nginx image validator

	// reg := regexp.MustCompile(`(?m)(nginx|nginx:\S)`)
	// for _, c := range pod.Spec.Containers {
	// 	fmt.Println("Image Name: " + c.Image)
	// 	if !reg.MatchString(c.Image) {
	// 		response.Allowed = false
	// 		break
	// 	}
	// }

	saName := pod.Spec.ServiceAccountName
	namespace := pod.Namespace

	sa, sa_err := Clientset.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), saName, metav1.GetOptions{})
	if sa_err != nil {
		fmt.Println("Can't find Service Account " + saName)
		fmt.Println(sa_err)
	}

	annotations := sa.Annotations
	doesHaveAWSAnnotation := false
	for key, val := range annotations {
		if strings.EqualFold(key, SA_Annotation) && strings.HasPrefix(val, Env_value) {
			doesHaveAWSAnnotation = true
			break
		}
	}
	Reset()
	if doesHaveAWSAnnotation {
		for _, containers := range pod.Spec.Containers {

			//checking for pod envs
			envs := containers.Env
			c := 0
			for _, env := range envs {
				if strings.EqualFold(env.Name, Env_name) && strings.HasPrefix(env.Value, Env_value) {
					fmt.Printf("Env => %s : %s\n", Env_name, env.Value)
					c++
					break
				}
			}
			if c == 1 {
				fmt.Printf("Container Name = %s\n", containers.Name)
				RecordValidationGauge(namespace, pod.Name, "Mutated")
			} else {
				RecordValidationFailures()
				RecordValidationGauge(namespace, pod.Name, "Failed")
			}
		}
	} else {
		RecordValidationGauge(namespace, pod.Name, "Ignored")
	}

	fmt.Println("Validated")
	responseAR := &admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	}

	json.NewEncoder(w).Encode(responseAR)
}

func handleErr(w http.ResponseWriter, r *admission.AdmissionReview, err error) {
	if err != nil {
		fmt.Println("Error Happened")
		log.Println(err)
	}
	response := &admission.AdmissionResponse{
		Allowed: true,
	}

	if r != nil {
		response.UID = r.Request.UID
		RecordRequests("POST", "500")
	} else {
		RecordRequests("POST", "204")
	}

	//r.Response = response
	responseAR := &admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	}
	json.NewEncoder(w).Encode(responseAR)
}
