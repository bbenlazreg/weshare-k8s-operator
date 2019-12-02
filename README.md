# weshare-k8s-operator

```bash
operator-sdk build gcr.io/gcp-project-name/weshare-operator
gcloud docker -- push gcr.io/gcp-project-name/weshare-operator:latest
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/crds/app.wescale.fr_appsets_crd.yaml
kubectl create -f deploy/operator.yaml
kubectl create -f deploy/crds/app.wescale.fr_v1alpha1_appset_cr.yaml
```
