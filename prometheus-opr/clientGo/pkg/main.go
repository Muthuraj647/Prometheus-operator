package main

import (
	"context"
	goflag "flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	prometheus "github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	flags "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

//for prometheus
var mutated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "admission_controller_pod_mutate_status",
	Help: "Pods Mutated by Admission Controller",
},
	[]string{"namespace", "pod", "container", "mutatingservice", "status"},
)

var mutate_ignored = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "admission_controller_pod_mutate_ignored_reason",
	Help: "Pods not Mutated by Admission Controller",
},
	[]string{"namespace", "pod", "container", "reason"},
)

func register() {
	prometheus.MustRegister(mutated)
	prometheus.MustRegister(mutate_ignored)
}
func init() {
	//registering the metrics
	register()
}

//record the metrics data points
func record(namespace, pod, container, mutatingservice, reason, status string, ignored bool) {

	if ignored {
		mutate_ignored.WithLabelValues(namespace, pod, container, reason).Add(1)
	} else {
		mutated.WithLabelValues(namespace, pod, container, mutatingservice, status).Add(1)
	}

}

//for setting prometheus endpoint
func operation(intervalOfListing int) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		fmt.Println("<<<---List--->>>")
		for _, podInfo := range (*pods).Items {
			t := podInfo.Status.StartTime
			t1 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)

			age := time.Since(t1)

			//logic for filter pods started 2 mins before

			if age.Minutes() <= 1 {
				namespace := podInfo.Namespace
				name := podInfo.Name

				fmt.Printf("Pod Name = %s\n", name)
				fmt.Printf("Pod Namespace = %s\n", namespace)
				fmt.Printf("Pod Status = %s\n", podInfo.Status.Phase)
				fmt.Printf("Age = %s\n", age)

				//to read service account
				saName := podInfo.Spec.ServiceAccountName

				//find sa from API server

				sa, sa_err := clientset.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), saName, metav1.GetOptions{})
				if sa_err != nil {
					fmt.Println("Can't find Service Account " + saName)
				}

				annotations := sa.Annotations
				doesHaveAWSAnnotation := false
				for key, val := range annotations {
					if strings.EqualFold(key, "eks.amazonaws.com/role-arn") && strings.HasPrefix(val, "arn:aws:iam::") {
						doesHaveAWSAnnotation = true
						break
					}
				}
				//to get particular namespace

				ns, _ := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})

				labels := ns.Labels

				var sidecarEnabled = false
				for key, val := range labels {
					if key == "istio-injection" && val == "enabled" {
						sidecarEnabled = true
						break
					}
				}
				sidecar := 0

				if sidecarEnabled || doesHaveAWSAnnotation {

					for _, containers := range podInfo.Spec.Containers {
						if sidecarEnabled && sidecar == 0 {
							if strings.EqualFold(containers.Name, "istio-proxy") {
								fmt.Println("sidecar injected")
								sidecar = 1
								continue
							}
						}

						if doesHaveAWSAnnotation {

							//checking for pod envs
							envs := containers.Env
							env_var := "AWS_ROLE_ARN"
							c := 0
							for _, env := range envs {
								if strings.EqualFold(env.Name, env_var) && env.Value != "" {
									fmt.Printf("Env => %s : %s\n", env_var, env.Value)
									c++
									break
								}
							}
							if c == 1 {
								fmt.Printf("Container Name = %s\n", containers.Name)
								record(namespace, name, containers.Name, "Pod-Identity-Webhook", "", "true", false)
							} else {
								record(namespace, name, containers.Name, "Pod-Identity-Webhook", "", "false", false)
							}

						}
					}
				}

				if !sidecarEnabled {
					fmt.Println("side car injection disabled")
					record(namespace, name, "", "", "istio-injection:disabled", "", true)
				}

				if !doesHaveAWSAnnotation {
					fmt.Println("don't have aws-role-arn annotation")
					record(namespace, name, "", "", "Service-Account does not Annotated with aws-role-arn", "", true)
				}

				if sidecar == 1 {
					record(namespace, name, "istio-proxy", "sidecar-injector", "", "true", false)
				} else if sidecar == 0 && sidecarEnabled {
					record(namespace, name, "istio-proxy", "sidecar-injector", "", "false", false)
				}

			}
		}

		time.Sleep(time.Duration(intervalOfListing) * time.Second)
	}
}

func main() {

	//env := flags.String("env-name", "AWS_ROLE_ARN", "Env Variable Name to check")
	interval := flags.Int("interval", 60, "Interval of Pod Listing in Seconds, Default 60s")
	klog.InitFlags(goflag.CommandLine)

	goflag.CommandLine.VisitAll(func(f *goflag.Flag) {
		flags.CommandLine.AddFlag(flags.PFlagFromGoFlag(f))
	})

	flags.Parse()

	_ = goflag.CommandLine.Parse([]string{})
	//fmt.Printf("ENV args :%s\n", *env)
	go operation(*interval)
	//for setting prometheus endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8888", nil)

}
