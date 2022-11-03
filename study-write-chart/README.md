# 开发构建一个Helm Chart包
## 1. 创建一个chart

### 1.1 可以使用helm create chart_name的方式创建一个chart
```shell
helm create study-write-chart

study-write-chart/
├── charts
├── Chart.yaml
├── templates
│   ├── deployment.yaml
│   ├── _helpers.tpl
│   ├── ingress.yaml
│   ├── NOTES.txt
│   ├── serviceaccount.yaml
│   ├── service.yaml
│   └── tests
│       └── test-connection.yaml
└── values.yaml
```
### 1.2 chart目录结构解释
* Chart.yaml：用于描述这个chart的基本信息，包括名字、描述信息、版本信息等。

* values.yaml：用于存储templates目录中模板文件中用到的变量信息，也就是说template中的模板文件引用的是values.yaml中的变量。

* templates：用于存放部署使用的yaml文件模板，这里面的yaml都是通过各种判断、流程控制、引用变量去调用values中设置的变量信息，最后完成部署。

  deployment.yaml：deployment资源yaml文件。

  ingress.yaml：ingress资源文件。

  NOTES.txt：用于接收chart的帮助信息，helm install部署完成后展示给用户，也可以使用helm status列出信息。

  _helpers.tpl：模板助手文件，定义的值可在模板中使用，可以在整个chart中重复使用。

### 1.3 自定义一个chart包的流程
* 创建一个chart包。
* 将部署服务用到的yaml文件全部放到templates目录中，然后将yaml中可能每次都需要变动的地方修改为变量。
* 将每次都需要变动的地方写到values.yaml中，让模板文件去引用，即可完成部署。

## 2. Chart包构建的相关命令
* 创建chart：helm create chart_name

* 将chart打包：helm package chart_path

* 查看chart中yaml模板文件渲染信息：helm get manifest release_name

## 3. 自定义开发构建一个Helm Chart包
手动创建一个chart包部署一个web项目。
具体实现步骤：

​ 1.创建一个chart包结构目录。

​ 2.删除template下的所有文件。

​ 3.将之前通过yaml部署的web程序的yaml文件放到template目录中，然后将yaml中经常需要修改的参数用变量替代。

​ 4.编写values.yaml存放变量信息。

​ 5.部署chart。

### 3.1 创建一个Chart包
我们可以先手动创建一个chart，然后将其部署，观察默认创建的chart有什么配置信息。
```shell
1.创建chart
[/home/king/gowork/github.com/wongearl/study-write-chart/]# helm create test-app
Creating mychart

2.运行chart
[/home/king/gowork/github.com/wongearl/study-write-chart/]# helm install test-app test-app/
NAME: test-app
LAST DEPLOYED: Thu Nov  3 11:27:46 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=test-app,app.kubernetes.io/instance=test-app" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace default $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl --namespace default port-forward $POD_NAME 8080:$CONTAINER_PORT


3.查看创建的chart中values.yaml文件内容了解部署的是什么服务
[/home/king/gowork/github.com/wongearl/study-write-chart/]# cat test-app/values.yaml 
·······
replicaCount: 1
image:
  repository: nginx
  pullPolicy: IfNotPresent
········
#可以看到原来创建的chart没有进行过任何配置，默认是一个nginx容器

4.验证容器服务是否可用
[/home/king/gowork/github.com/wongearl/study-write-chart/]# kubectl get pod -o wide
test-app-59f967cdcb-w742s       1/1     Running   0          2m11s   10.244.1.96    node1   <none>           <none>
[/home/king/gowork/github.com/wongearl/study-write-chart/]# curl -I 10.244.1.96
HTTP/1.1 200 OK
Server: nginx/1.16.0
Date: Fri, 23 Jul 2021 02:55:57 GMT
Content-Type: text/html
Content-Length: 612
Last-Modified: Tue, 23 Apr 2019 10:18:21 GMT
Connection: keep-alive
ETag: "5cbee66d-264"
Accept-Ranges: bytes
```
### 3.2 自定义templates模板文件
```shell
templates下的模板文件中需要变动的信息都需要通过变量的方式调用。

调用values.yaml中变量的书写格式：{{ .Values.变量名 }}。
```
将templates目录下的yaml默认模板文件清空
```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/]# cd test-app/
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/]# rm -rf templates/*
```
#### 3.2.1 生成deployment资源模板文件
首先进入templates目录。
```shell
1.通过kubectl create命令生成一个deployment资源的yaml文件
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/templates]# kubectl create deployment test-app --image=nginx:1.16 -o yaml --dry-run=client > deployment.yaml

2.修改deployment.yaml文件将需要动态变动的信息替换成变量
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/templates]# vim deployment.yaml 
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: {{ .Values.appname }}			#将values.yaml中的appname对应的变量值渲染在这里
  name: test-app
spec:
  replicas: {{ .Values.replicas }}		#将values.yaml中的replicas对应的变量值渲染在这里
  selector:
    matchLabels:
      app: {{ .Values.appname }}		#标签可以和资源名称一样，因此也可以直接调用appname变量
  template:
    metadata:
      labels:
        app: {{ .Values.appname }}		#标签可以和资源名称一样，因此也可以直接调用appname变量
    spec:
      containers:
      - image: {{ .Values.image }}:{{ .Values.imageTag }}		#将values.yaml中的image、imageTag对应的变量值渲染在这里,表示镜像的版本号
        name: {{ .Values.appname }}			#容器的名称也和资源的名称保持一致即可
        ports:
        - name: web
          containerPort: 80
          protocol: TCP
        volumeMounts:
        - name: code
          mountPath: /data/code/
        - name: config
          mountPath: /data/nginx/conf/conf.d/
      volumes:	
        - name: config
          configMap:
            name: {{ .Values.appname }}-cm				#confimap的名字也可以使用程序名称的变量加上-cm
        - name : code
          persistentVolumeClaim:
            claimName: {{ .Values.appname }}-pvc		#pvc的名字也可以使用程序名称的变量加上-pv
            readOnly: false        
```
#### 3.2.2 生成service资源模板文件
首先进入templates目录。
```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/templates]# cat service.yaml 
apiVersion: v1
kind: Service
metadata:
  labels:
    app: {{ .Values.appname }}			#service要管理deployment的pod资源，因此这里的标签要和pod资源的标签对应上，直接调用appname这个变量
  name: {{ .Values.appname }}-svc		#service资源的名称，也可以直接调用appname这个变量，后面加一个-svc
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80 
  selector:
    app: {{ .Values.appname }}			#标签选择器还是调用appname这个变量
  type: NodePort
```
#### 3.2.4.生成configmap资源模板文件
生成一个cm资源，用于配置nginx找到test-app项目。
```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/templates]# vim configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.appname }}-cm			#引入appname变量加上-cm作为cm资源的名称
data:
  test.app.com.conf: |
    server {
      listen 80;
      server_name test.app.com;
      location / {
        root /data/code/test_app;
        index index.html;
      }
    }
  
```
#### 3.2.5.生成pv和pvc的资源模板文件
pvc存储web程序的代码

```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/templates]# vim pv-pvc.yaml
apiVersion: v1
kind:  PersistentVolume
metadata:
  name: {{ .Values.appname }}-pv			#引入appname变量加上-pv作为pv资源的名称
  labels:
    pv: {{ .Values.appname }}-pv			#标签也可以使用和pv名称一样的名字
spec:
  capacity:
    storage: 1Gi
  accessModes:
  - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  nfs:
    path: {{ .Values.nfsPath }}
    server: {{ .Values.nfsServer }}
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ .Values.appname }}-pvc			#引入appname变量加上-pvc作为pvc资源的名称
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
  selector:
    matchLabels:
      pv: {{ .Values.appname }}-pv			#指定pv的标签
```
### 3.3.自定义values变量文件
刚刚在资源模板文件中调用了很多变量，这一部将变量对应的值卸载values.yaml中。
```shell
chart根目录去写values.yaml
[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/]# pwd
/home/king/gowork/github.com/wongearl/study-write-chart/test-app/

[/home/king/gowork/github.com/wongearl/study-write-chart/test-app/]# vim values.yaml 
appname: test-app
replicas: 3
image: registry.cn-hangzhou.aliyuncs.com/earl-k8s/nginx
imageTag: 1.23.1
nfsPath: /data2/k8s/know
nfsServer: 192.168.75.130

#整个目录结构
[/home/king/gowork/github.com/wongearl/study-write-chart/]# tree test-app/
test-app/
├── charts
├── Chart.yaml
├── templates
│	 ├── configmap.yaml
│ 	 ├── deployment.yaml
│	 ├── pv-pvc.yaml
│	 └── service.yaml
└── values.yaml

```
### 3.4 部署自定义构建的Chart包
```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/]# helm install test-app test-app/
NAME: test-app
LAST DEPLOYED: Thu Nov  3 11:50:59 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```
### 3.5 查看变量渲染后的YAML模板文件
```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/]# helm get manifest test-app 
---
# Source: test-app/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-app-cm                     #引入appname变量加上-cm作为cm资源的名称
data:
  test.app.com.conf: |
    server {
      listen 80;
      server_name test.app.com;
      location / {
        root /data/code/test_app;
        index index.html;
      }
    }
---
# Source: test-app/templates/pv-pvc.yaml
apiVersion: v1
kind:  PersistentVolume
metadata:
  name: test-app-pv                     #引入appname变量加上-pv作为pv资源的名称
  labels:
    pv: test-app-pv                     #标签也可以使用和pv名称一样的名字
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  nfs:
    path: /data2/k8s/know
    server: 192.168.75.130
---
# Source: test-app/templates/pv-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-app-pvc                    #引入appname变量加上-pvc作为pvc资源的名称
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
  selector:
    matchLabels:
      pv: test-app-pv                   #指定pv的标签
---
# Source: test-app/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-app                       #service要管理deployment的pod资源，因此这里的标签要和pod资源的标签对应上，直接调用appname这个变量
  name: test-app-svc            #service资源的名称，也可以直接调用appname这个变量，后面加一个-svc
spec:
  ports:
    - port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: test-app                       #标签选择器还是调用appname这个变量
  type: NodePort
---
# Source: test-app/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-app                       #将values.yaml中的appname对应的变量值渲染在这里
  name: test-app
spec:
  replicas: 3           #将values.yaml中的replicas对应的变量值渲染在这里
  selector:
    matchLabels:
      app: test-app             #标签可以和资源名称一样，因此也可以直接调用appname变量
  template:
    metadata:
      labels:
        app: test-app           #标签可以和资源名称一样，因此也可以直接调用appname变量
    spec:
      containers:
        - image: registry.cn-hangzhou.aliyuncs.com/earl-k8s/nginx:1.23.1                #将values.yaml中的image、imageTag对应的变量值渲染在这里,表示镜像的版本号
          name: test-app                        #容器的名称也和资源的名称保持一致即可
          ports:
            - name: web
              containerPort: 80
              protocol: TCP
          volumeMounts:
            - name: code
              mountPath: /data/code/
            - name: config
              mountPath: /data/nginx/conf/conf.d/
      volumes:
        - name: config
          configMap:
            name: test-app-cm                           #confimap的名字也可以使用程序名称的变量加上-cm
        - name : code
          persistentVolumeClaim:
            claimName: test-app-pvc             #pvc的名字也可以使用程序名称的变量加上-pv
            readOnly: false
```
可以清楚的看到完整的yaml文件，在模板文件中的变量以及替换成了values的变量值。
## 4. 将Chart包打包
```shell
[/home/king/gowork/github.com/wongearl/study-write-chart/]# helm  package test-app/
Successfully packaged chart and saved it to: /home/king/gowork/github.com/wongearl/study-write-chart/test-app-0.1.0.tgz

[/home/king/gowork/github.com/wongearl/study-write-chart/]# ll
drwxr-xr-x  4 king king  4096 11月  3 11:50 test-app/
-rw-rw-r--  1 king king  1816 11月  3 11:54 test-app-0.1.0.tgz
```

