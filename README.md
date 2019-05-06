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

3. Install database schema to Cassandra:
    ```
    kubectl run cqlsh --image=cassandra:latest --generator=run-pod/v1 --restart=Never -- cqlsh cassandra-0.cassandra.default.svc.cluster.local -e  "CREATE KEYSPACE caas WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 3}; use caas; CREATE TABLE counter (name text, value counter, PRIMARY KEY (name))"
    ```
    Check that pod cqlsh finished successfully. It does not log anything on success.

4. For debugging, check Cassandra node status:
    ```
    $ kubectl exec cassandra-0 nodetool status
    
    ```
5. Run the app (TODO: add service and route)
    ```
    kubectl create -f caas.yaml
    ```

6. Test the application

    ```
    $ kubectl get pod -o wide
    NAME                   READY   STATUS    RESTARTS   AGE   IP           NODE        NOMINATED NODE   READINESS GATES
    caas-b4b67f497-82c8h   1/1     Running   0          4s    172.17.0.5   127.0.0.1   <none>           <none>

    $ curl http://172.17.0.5/first/json
    {"name":"first","count":1}
    $ curl http://172.17.0.5/first/json
    {"name":"first","count":2}
    $ curl http://172.17.0.5/first/json
    {"name":"first","count":3}

    $ curl http://172.17.0.5/second/json
    {"name":"second","count":1}
    $ curl http://172.17.0.5/second/json
    {"name":"second","count":2}

    ## http://172.17.0.5/<counter name>/json

7. Scale up
    ```
    $ kubectl scale --replicas=4 statefulset/cassandra
    $ kubectl get pod
    ```
    * Database rows were redistributed to the new node:
        ```
        $ kubectl exec cassandra-0 nodetool status
        ```

8. Simulate "graceful" failure
    ```
    kubectl delete pod cassandra-1
    kubectl get pod -w
    ```
