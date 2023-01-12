package k8sunit

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimefake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

var (
	// Scheme contains all types of custom clientset and kubernetes client-go clientset
	Scheme = runtime.NewScheme()
)

func init() {
	// 添加你需要的资源的 scheme
	//_ = clientgoscheme.AddToScheme(Scheme)
	//_ = cosscheme.AddToScheme(Scheme)
	//_ = apiextensionsv1.AddToScheme(Scheme)
}

// 原生 client
// 指原生的 client-go 和使用 code-generator 生成的各 CR 的 typed client，这些 client 都提供了相应的 fake client 方法
func TestAdd(t *testing.T) {
	tests := []struct {
		name             string
		objects          []runtime.Object
		event            *corev1.Event
		isErr            bool
		wantedEventCount int32
	}{
		{
			name:             "exist test",
			objects:          []runtime.Object{GetEvent()},
			event:            GetEvent(),
			isErr:            false,
			wantedEventCount: 1,
		},
		{
			name:             "not exist test",
			event:            GetEvent(),
			isErr:            false,
			wantedEventCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成 fake client，把需要 CRUD 的资源加入进去
			client := fake.NewSimpleClientset(tt.objects...)
			// 创建基于 fake client 的 informer
			informer := informers.NewSharedInformerFactory(client, 0)

			// 往 informer indexer 中添加对应的资源对象
			events := []corev1.Event{}
			for _, s := range events {
				informer.Events().V1().Events().Informer().GetIndexer().Add(s)
			}
			// 执行函数
			err := Add(client, tt.event)
			if tt.isErr != (err != nil) {
				t.Errorf("%s Add() unexpected error: %v", tt.name, err)
			}

			// 校验结果是否符合预期
			eventObj, err := client.CoreV1().Events(tt.event.Namespace).Get(context.TODO(), tt.event.Name, metav1.GetOptions{})
			if err != nil {
				t.Errorf("%s unexpected error: %v", tt.name, err)
			}
			if eventObj.Count != tt.wantedEventCount {
				t.Errorf("%s event Count = %d, want %d", tt.name, eventObj.Count, tt.wantedEventCount)
			}
		})
	}

	// generic client
	//指 controller-runtime 提供的 generic client，和上面的 typed client 不同的是，该 client 是一个通用的 client，可用于 CRUD 任何资源，测试方法基本和上面一样，只是构造 fake client 方法稍有不同，其它流程都一样，下面就只介绍构造 fake client 的方法，其它的内容就不再赘述了。
	//构造 generic client 的 fake client 的方法最大的一个不同点是它需要包含要 CRUD 的所有资源的 scheme：
	// 配置 scheme 和需要添加的 objects
	client := runtimefake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&corev1.Event{}).Build()
	fmt.Println(client)

}

func TestCreate(t *testing.T) {
	//// 生成 env，这里面有很多配置项，可以根据需要配置
	//testEnv := &envtest.Environment{}
	//
	//// 启动环境，返回值是环境的 rest.Config，可用于生成各 kube client
	//config, err := testEnv.Start()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//// 销毁集群
	//testEnv.Stop()
	// 启动测试集群
	testEnv := &envtest.Environment{}
	config, err := testEnv.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer testEnv.Stop()

	// 生成 generic client
	cli, err := client.New(config, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		t.Fatal(err)
	}

	// 准备测试数据
	sc1 := GetStorageClass()
	sc1.Name = "sc1"
	sc1.Provisioner = "example.com/test"
	// 创建测试数据
	if err := cli.Create(context.TODO(), sc1); err != nil {
		t.Fatal(err)
	}

	// 开始执行测试...

	type args struct {
		client clientset.Interface
		crds   []*apiextensionsv1.CustomResourceDefinition
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Create(tt.args.client, tt.args.crds...); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
