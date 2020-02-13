# Kudecs
<a href="https://hub.docker.com/repository/docker/kubernetesmisc/kudecs"><img src="https://img.shields.io/badge/Docker-v1.0.0-366934"></a>&nbsp;
<img src="https://img.shields.io/badge/Version-v1.0.0-366934">&nbsp;
<a href="https://goreportcard.com/report/github.com/kubernetes-misc/kudecs"><img src="https://goreportcard.com/badge/github.com/kubernetes-misc/kudecs"></a>&nbsp;
<a href="https://codebeat.co/projects/github-com-kubernetes-misc-kudecs-master"><img alt="codebeat badge" src="https://codebeat.co/badges/482ac388-fd64-4e9a-9dcd-f4b280889ad4" /></a>&nbsp;
<a href="https://codeclimate.com/github/kubernetes-misc/kudecs/maintainability"><img src="https://api.codeclimate.com/v1/badges/5930e15ac6ea7c033eb6/maintainability" /></a>


KUberenetes DEclarative Certificates as Secrets<br />

## How to use

### Create the namespace, CRD, service account, cluster role, cluster role binding and deployment
```shell script
kubectl create ns kudecs
kubectl apply -n kudecs -f https://raw.githubusercontent.com/kubernetes-misc/kudecs/master/install/crd.yaml
kubectl apply -n kudecs -f https://raw.githubusercontent.com/kubernetes-misc/kudecs/master/install/clusterrole.yaml
kubectl apply -n kudecs -f https://raw.githubusercontent.com/kubernetes-misc/kudecs/master/install/deployment.yaml
```

Check that you can see the changes
```shell script
kubectl get all -n kudecs
```


### Create a kudec CRD
```shell script
kubectl apply -f https://raw.githubusercontent.com/kubernetes-misc/kudecs/master/examples/tiny-example.yaml
```

Check that you can see the changes
```shell script
kubectl get kudecs -o yaml
```

Check the application logs to see what kudecs says about your kudec
```shell script
kubectl logs deployment/kudecs
```

If everything went well you will be able to see secrets. Otherwise check the steps above.
```shell script
kubectl get secrets -o yaml
```


## Roadmap

### Version 2
- HA support, deployment
- Namespace inclusions / exclusions
- Delete policy (managed / master-only)
- External certificate generation providers




