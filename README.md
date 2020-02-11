# Kudecs
<img src="https://img.shields.io/badge/Version-v0.2.0-f5bc42">&nbsp;
<a href="https://goreportcard.com/report/github.com/kubernetes-misc/kudecs"><img src="https://goreportcard.com/badge/github.com/kubernetes-misc/kudecs"></a>&nbsp;
<a href="https://codebeat.co/projects/github-com-kubernetes-misc-kudecs-master"><img alt="codebeat badge" src="https://codebeat.co/badges/482ac388-fd64-4e9a-9dcd-f4b280889ad4" /></a>&nbsp;
<a href="https://codeclimate.com/github/kubernetes-misc/kudecs/maintainability"><img src="https://api.codeclimate.com/v1/badges/5930e15ac6ea7c033eb6/maintainability" /></a>


KUberenetes DEclarative Certificates as Secrets<br />

## How to use

### Create the crd
```shell script
kubectl apply -f https://raw.githubusercontent.com/kubernetes-misc/kudecs/master/yaml/crd.yaml
```

### Deploy to cluster
- Create a new file called `kudecs.yaml`
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kudecs
  namespace: kudecs
spec:
  selector:
    matchLabels:
      app: kudecs
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: kudecs
    spec:
      containers:
      - env:
        - name: storeNamespace
          value: kudecs
        image: kubernetesmisc/kudecs:latest
        imagePullPolicy: Always
        name: kudecs
        resources:
          limits:
            cpu: 500m
            memory: 64Mi
          requests:
            cpu: 500m
            memory: 64Mi
```


### Create a kudec (CRD)

- Create a new kudecs yaml file `example.yaml` and populate with the following yaml
```yaml
apiVersion: "kubernetes-misc.xyz/v1"
kind: kudec
metadata:
  name: example
spec:
  days: 365
  countryName: USA
  stateName: NY
  organizationName: Kubernetes Misc
  organizationalUnit: IT
  injectedSecrets:
    - namespace: default
      secretName: okwp-secret-key
      sourceKey: private
      keyName: private
    - namespace: default
      secretName: okwp-secret-pub
      sourceKey: public
      keyName: public
```

Apply the example
```shell script
kubectl apply -f example.yaml
```

Check that you can see your change
```shell script
kubectl get kudecs/example -o yaml
```

Check the application logs to see that it is running
```shell script
kubectl logs deployment/kudecs
```

If everything went well you will be able to see secrets. Otherwise check the steps above.
```shell script
kubectl get secrets -o yaml
```


## Roadmap

### Version 1
- In-cluster authentication
- Official :latest docker image
- Deployment yaml

### Version 2
- HA deployment support
- Namespace restrictions / exclusions




