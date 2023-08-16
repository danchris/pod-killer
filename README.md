# pod-killer

pod-killer monitors the pods containing the `pod-killer/name` and `pod-killer/alive` and its goal is to delete the unnecessary pods so that only as many pods as are declared in the `pod-killer/alive` to be running.

## Example

As you can see in the example below, the declared replicas are 2 but the `pod-killer/alive: "1"`. So, pod-killer will delete one replica of the `nginx` deployment.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: nginx
  name: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
      pod-killer/name: "nginx"
      pod-killer/alive: "1"
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: nginx
        pod-killer/name: "nginx"
        pod-killer/alive: "1"
    spec:
      containers:
      - image: nginx
        name: nginx
```
