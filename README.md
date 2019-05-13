# CaaS
Counter as a service (simple cassandra demo).

Usage:

1. Install Cassandra Service and StatefulSet (from https://kubernetes.io/docs/tutorials/stateful-application/cassandra/)
    ```
    kubectl create -f https://raw.githubusercontent.com/jsafrane/caas/master/cassandra-service.yaml
    kubectl create -f https://raw.githubusercontent.com/jsafrane/caas/master/cassandra-statefulset.yaml
    ```
    
    Changes from Kubernetes docs:
    * No StorageClass is created. Upstream example targets minikube, we run on GCE with a default storage class.
    * Lower CPU requests. We want to run the demo on a 3 node cluster and we don't require extra speed.

2. Wait for Cassandra pods to be up (~6 minutes !!!)

3. For debugging, check Cassandra node status:
    ```
    $ kubectl exec cassandra-0 nodetool status
    
    ```

4. Run the app
    ```
    kubectl create -f https://raw.githubusercontent.com/jsafrane/caas/master/caas.yaml
    ```

5. Test the application

    ```
    $ kubectl get service
    NAME         TYPE           CLUSTER-IP   EXTERNAL-IP   PORT(S)        AGE
    caas         LoadBalancer   10.0.0.2     <pending>     80:31024/TCP   3m36s

    $ curl http://10.0.0.2/first/html
    <html>
      <head><title>first</title></head>
    <body>
    <h1>Counter: first, value: 6</h1>
    ...
    ```

    ## http://10.0.0.2/<counter name>/html

6. Scale up
    ```
    $ kubectl scale --replicas=4 statefulset/cassandra
    $ kubectl get pod
    ```
    * Database rows were redistributed to the new node:
        ```
        $ kubectl exec cassandra-0 nodetool status
        ```

7. For debugging, check the database content
    ```
    $ kubectl run -ti cqlsh --image=cassandra:latest --generator=run-pod/v1 cqlsh cassandra-0.cassandra.default.svc.cluster.local
    [cqlsh 5.0.1 | Cassandra 3.11.2 | CQL spec 3.4.4 | Native protocol v4]
    Use HELP for help.
    cqlsh> use caas;
    cqlsh:caas> select * from counter;
    
     name  | value
    -------+-------
     first |     6
    
    (1 rows)
    ```

8. Simulate "graceful" failure
    ```
    kubectl delete pod cassandra-1
    kubectl get pod -w
    ```
