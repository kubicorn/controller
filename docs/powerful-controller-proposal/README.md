# Powerful Controller Proposal

In this proposal the controller (by design) will have read/write permissions to it's own `CRD`.

## Entry point

The controller will accept a Kubicorn `cluster.Cluster{}` and will bootstrap itself.
This can come from anywhere, as long as we pass it into the main controller function at Runtime.

## Bootstrap

#### 1) Deploy self to Kubernetes

All of the information needed to authenticate `client-go` against Kubernetes lives in the `cluster.Cluster{}` passed into the controller.
We authenticate with Kubernetes and then deploy a copy of our self to Kubernetes.

 - Which namespace?
 - Which resources?

#### 2) Ensure the CRDs are defined

After the controller is running in Kubernetes, the first thing we do is ensure the `CRD`s are populated from the data in `cluster.Cluster{}` passed into the controller.

#### 3) Delete the internal memory cache

After the `CRD`s are populated, we can destroy our `cluster.Cluster{}` cache

#### 4) Start the control loop

We then start a control loop that completes the following steps indefinitely

 - Read from the `CRD`
 - Detect if *anything* has changed (via the `kubicorn/kubicorn/pkg/compare` package), otherwise we no-op
 - If a change is detected we re-build the `cluster.Cluster{}` from the `CRD`s
 - We then call the Atomic Reconciler and reconcile the newly created `cluster.Cluster{}`
 - We update the `CRD` with a status, and an updated at time
