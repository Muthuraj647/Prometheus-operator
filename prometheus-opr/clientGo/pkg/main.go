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
var eksmutated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "eks_pod_identity_webhook_mutated",
	Help: "Number of Pods Mutated with EKS Pod Identity webhook",
},
	[]string{"namespace", "pod", "container"},
)

func register() {
	prometheus.MustRegister(eksmutated)
}
func init() {
	//registering the metrics
	register()
}

//record the metrics data points
func record(namespace, pod, container string, mutated bool) {
	if mutated {
		eksmutated.WithLabelValues(namespace, pod, container).Add(1)
	} else {
		eksmutated.WithLabelValues(namespace, pod, container).Add(0)
	}
}

func main() {

	env := flags.String("env-name", "AWS_ROLE_ARN", "Env Variable Name to check")
	interval := flags.Int("interval", 60, "Interval of Pod Listing in Seconds, Default 60s")
	klog.InitFlags(goflag.CommandLine)

	goflag.CommandLine.VisitAll(func(f *goflag.Flag) {
		flags.CommandLine.AddFlag(flags.PFlagFromGoFlag(f))
	})

	flags.Parse()

	_ = goflag.CommandLine.Parse([]string{})
	fmt.Printf("ENV args :%s", *env)
	go operation(*env, *interval)
	//for setting prometheus endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8888", nil)

}

//for setting prometheus endpoint
func operation(env_var string, intervalOfListing int) {
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
				//fmt.Printf("Env => %s = %s\n", podInfo.Spec.Containers[0].Env.Name, podInfo.Spec.Containers[0].Env.Value)

				for _, containers := range podInfo.Spec.Containers {
					if strings.EqualFold(containers.Name, "Istio-Proxy") {
						continue
					}
					envs := containers.Env
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
						record(namespace, name, containers.Name, true)
					} else {
						record(namespace, name, containers.Name, false)
					}
				}

			}
		}

		time.Sleep(time.Duration(intervalOfListing) * time.Second)
	}
}
