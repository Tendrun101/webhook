
## 简述
修改新建pod的webhook，练习

## 部署

### 1.build 镜像
make

### 2.生成Secret以及yamls
./deploy/create-certs-secret.sh

### 3.部署相关对象
kubectl apply -f webhook-certs.yaml \
kubectl apply -f webhook-registration.yaml \
kubectl apply -f webhook.yaml