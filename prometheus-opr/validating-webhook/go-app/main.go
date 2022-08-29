package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	admission "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validation(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Println("Method Not Allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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

	reg := regexp.MustCompile(`(?m)(nginx|nginx:\S)`)
	for _, c := range pod.Spec.Containers {
		fmt.Println("Image Name: " + c.Image)
		if !reg.MatchString(c.Image) {
			response.Allowed = false
			break
		}
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
		Allowed: false,
	}

	if r != nil {
		response.UID = r.Request.UID
	}

	r.Response = response
	json.NewEncoder(w).Encode(r)
}

// func main() {
// 	fmt.Println("Ready to Validate")
// 	m := http.NewServeMux()
// 	m.HandleFunc("/", validation)
// 	http.ListenAndServe(":8089", m)
// }
