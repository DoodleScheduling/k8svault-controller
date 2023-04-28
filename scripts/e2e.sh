#!/bin/bash

echo "enable vault secret engine kv v1"
kubectl -n k8svault-system wait pods/vault-0 --for=condition=Ready --timeout=1m
kubectl -n k8svault-system exec -i sts/vault vault -- vault secrets disable kv
kubectl -n k8svault-system exec -i sts/vault vault -- vault secrets disable secret
kubectl -n k8svault-system exec -i sts/vault vault -- vault secrets enable -version=1 -path=/secret kv

echo "enable vault kubernetes authentication"
kubectl -n k8svault-system exec -i sts/vault vault -- vault auth enable kubernetes

echo 'path "secret/*" { capabilities = ["update","read","create"]  }' | kubectl exec -i -n k8svault-system vault-0 -- vault policy write k8svault-controller -
kubectl -n k8svault-system exec -i sts/vault vault -- vault write auth/kubernetes/role/k8svault-controller bound_service_account_names=default bound_service_account_namespaces=k8svault-system policies=k8svault-controller
k8s_host="$(kubectl exec vault-0 -n k8svault-system -- printenv | grep KUBERNETES_PORT_443_TCP_ADDR | cut -f 2- -d "=" | tr -d " ")"
k8s_port="443"
kubectl get serviceaccount vault -n k8svault-system -o yaml

k8s_cacert="$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 --decode)"
tr_account_token="$(kubectl get secret vault -n k8svault-system -o go-template='{{ .data.token }}' | base64 --decode)"

kubectl -n k8svault-system exec -i sts/vault vault -- vault write auth/kubernetes/config token_reviewer_jwt="${tr_account_token}" kubernetes_host="https://${k8s_host}:${k8s_port}" kubernetes_ca_cert="${k8s_cacert}"

kubectl -n k8svault-system wait vaultbinding/test-secret --for=condition=Bound --timeout=1m
kubectl -n k8svault-system exec -i sts/vault vault -- vault read /secret/example
kubectl -n k8svault-system wait vaultmirror/test-secret --for=condition=Bound --timeout=1m
kubectl -n k8svault-system exec -i sts/vault vault -- vault read /secret/example-copy
