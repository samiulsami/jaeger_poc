### [WIP] Test project for learning Jaeger/OTeL/Thanos

#### build and run
```bash
docker build -t sami7786/jaeger_poc:latest .
docker push sami7786/jaeger_poc:latest
docker run -p 8080:8080 sami7786/jaeger_poc:latest -d
```

```bash
kubectl create -f hack/

kubectl port-forward svc/go-apiserver 8080:80
kubectl port-forward svc/jaeger-ui -n observability 16686:80

curl http://localhost:8080/hello
```


