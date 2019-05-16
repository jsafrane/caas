# CaaS
Counter as a service (simple Apache Cassandra demo).

Usage:

1. Install Cassandra Service and StatefulSet.
    ```
    kubectl apply -f https://raw.githubusercontent.com/jsafrane/caas/master/cassandra.yaml
    ```
    
    Changes from [Kubernetes docs](https://kubernetes.io/docs/tutorials/stateful-application/cassandra/):
    * Reshuffled the YAML to have it in "talk friendly" order.
    * No StorageClass is created. Upstream example targets minikube, we run on GCE with a default storage class.
    * Lower CPU requests. We want to run the demo on a 3 node cluster and we don't require extra speed.

2. Wait for Cassandra pods to be up (~6 minutes !)
   * While it is starting, we can 'debug' it a bit:
   
    ```
    $ kubectl get pvc

    $ kubectl describe pvc

    $ kubectl get pod
    
    $ kubectl describe pod cassandra-0
    
    $ kubectl logs cassandra-0
    ```

4. Install the application.
    ```
    kubectl apply -f https://raw.githubusercontent.com/jsafrane/caas/master/caas.yaml
    ```

5. Test the application.

    ```
    $ kubectl get service
    NAME         TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)        AGE
    caas         LoadBalancer   10.43.247.108 35.238.128.239 80:31501/TCP   1m

    $ curl http://35.238.128.239/first/html
    <html>
      <head><title>first</title></head>
    <body>
    <h1>Counter: first, value: 6</h1>
    ...
    ```

    `http://35.238.128.239/<counter name>/html`

    For debugging, you can check Cassandra node status:
    ```
    $ kubectl exec cassandra-0 nodetool status
    Datacenter: DC1-K8Demo
    ======================
    Status=Up/Down
    |/ State=Normal/Leaving/Joining/Moving
    --  Address    Load       Tokens  Owns (effective)  Host ID                               Rack
    UN  10.40.0.8  136.15 KiB 32      100.0%            d55b2a33-34a7-4216-8a21-77606832d847  Rack1-K8Demo
    UN  10.40.2.8  115.63 KiB 32      100.0%            7ead59de-b0ed-4b7f-b1c0-52c0344ace51  Rack1-K8Demo
    UN  10.40.1.5  113.59 KiB 32      100.0%            0514d1a3-e0e8-4f42-8e3e-697cf958b8d6  Rack1-K8Demo
    ```

6. Scale up.
    ```
    $ kubectl scale --replicas=4 statefulset/cassandra
    $ kubectl get pod
    ```
    * Database rows were redistributed to the new node:
        ```
        $ kubectl exec cassandra-0 nodetool status
        ```

7. Simulate "graceful" failure.
    ```
    $ kubectl delete pod cassandra-1
    $ kubectl get pod -w
    ```

8. Scale down
    ```
    $ kubectl scale --replicas=3 statefulset/cassandra
    $ kubectl exec cassandra-0 nodetool status
    Datacenter: DC1-K8Demo
    ======================
    Status=Up/Down
    |/ State=Normal/Leaving/Joining/Moving
    --  Address    Load       Tokens Owns (effective)  Host ID                               Rack
    UN  10.40.0.8  141.08 KiB  32    78.6%             d55b2a33-34a7-4216-8a21-77606832d847  Rack1-K8Demo
    UN  10.40.2.8  120.56 KiB  32    86.7%             7ead59de-b0ed-4b7f-b1c0-52c0344ace51  Rack1-K8Demo
    UN  10.40.1.5  118.52 KiB  32    67.8%             0514d1a3-e0e8-4f42-8e3e-697cf958b8d6  Rack1-K8Demo
    DN  10.40.1.6  75.89 KiB   32    67.0%             9bfa542d-55ca-46b7-b109-aa10b477dfc0  Rack1-K8Demo
    
    $ kubectl exec cassandra-0 nodetool removenode 9bfa542d-55ca-46b7-b109-aa10b477dfc0
    
    $ kubectl exec cassandra-0 nodetool status
    Datacenter: DC1-K8Demo
    ======================
    Status=Up/Down
    |/ State=Normal/Leaving/Joining/Moving
    --  Address    Load       Tokens       Owns (effective)  Host ID                               Rack
    UN  10.40.0.8  141.08 KiB  32           100.0%            d55b2a33-34a7-4216-8a21-77606832d847  Rack1-K8Demo
    UN  10.40.2.8  135.4 KiB   32           100.0%            7ead59de-b0ed-4b7f-b1c0-52c0344ace51  Rack1-K8Demo
    UN  10.40.1.5  128.51 KiB  32           100.0%            0514d1a3-e0e8-4f42-8e3e-697cf958b8d6  Rack1-K8Demo
    ```

## Debugging tools

* DNS
    * Run nslookup in a pod in the cluster:
    ```shell
    $ kubectl run --restart=Never --rm -ti --generator=run-pod/v1 nslookup --image=tutum/dnsutils nslookup cassandra.default.svc.cluster.local   
    ```

* CQL
    * Run CQL query in Cassandra:
    ```shell
    $ kubectl run --restart=Never --rm -ti --generator=run-pod/v1 cqlsh --image=cassandra:latest -- cqlsh cassandra-0.cassandra.default.svc.cluster.local -k caas
    (wait a bit, a large image is pulled)
    cqlsh:caas> select * from counter;
    
    name   | value
    --------+-------
    sample |    51
