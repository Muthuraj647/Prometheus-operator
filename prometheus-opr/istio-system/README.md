Istio

To install istio:

1. Download the files from istio:
    curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.14.3 TARGET_ARCH=x86_64 sh -

2. Move to the Istio package directory. 
    cd istio-1.14.3

3. Add the istioctl client to your path 
    export PATH=$PWD/bin:$PATH

4. install with Demo Profile:
    istioctl install --set profile=demo -y

5. Add a namespace label to instruct Istio to automatically inject Envoy sidecar proxies when you deploy your application later:
    kubectl label namespace default istio-injection=enabled


6. To install Book Info Sample Application:
    kubectl apply -f samples/bookinfo/platform/kube/bookinfo.yaml

7. To install gateway

    cd K8s/Prometheus-operator/Prometheus-operator/prometheus-opr/istio-system
    kubectl apply -f gateway.yaml

8. Ensure that there are no issues with the configuration
    istioctl analyze