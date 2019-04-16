# CaaS
Counter as a service (simple cassandra demo).

Usage:

1. Install Cassandra service and statefulset (from https://kubernetes.io/docs/tutorials/stateful-application/cassandra/)
    ```
    kubectl create -f cassandra-service.yaml
    kubectl create -f cassandra-statefulset.yaml
    ```

2. Wait for Cassandra pods to be up (~6 minutes !!!)

3. Install database schema to Cassandra (TODO: add to the application?)
    ```
    kubectl run cqlsh --image=cassandra:latest --generator=run-pod/v1 --restart=Never -- cqlsh cassandra-0.cassandra.default.svc.cluster.local -e  "CREATE KEYSPACE caas WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 3}; use caas; CREATE TABLE counter (name text, value counter, PRIMARY KEY (name))"
    ```
    Check that pod cqlsh finished successfully. It does not log anything on success.

4. Run the app (TODO: add service and route)
    ```
    kubectl create -f caas.yaml
    ```

5. Test the application

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


TODO: test with some pods down, scale up/down etc.
