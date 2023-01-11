// Copyright 2017 EasyStack, Inc.
package handlers

import (
	"github.com/linkall-labs/vanus-operator/api/models"
	"github.com/linkall-labs/vanus-operator/api/restapi/operations/cluster"

	"github.com/go-openapi/runtime/middleware"
)

var (
	WorkloadOldReplicAnno = "replic.ecns.io/workload"
	// revisionAnnoKey       = "deployment.kubernetes.io/revision"
	// rolloutReasonAnno     = "kubernetes.io/change-cause"
)

// 所有注册的处理函数 应该 按照顺序出现在 Registxxx 下

func RegistClusterHandler(a *Api) {
	a.ClusterCreateClusterHandler = cluster.CreateClusterHandlerFunc(a.createClusterHandler)
	a.ClusterDeleteClusterHandler = cluster.DeleteClusterHandlerFunc(a.deleteClusterHandler)
	a.ClusterPatchClusterHandler = cluster.PatchClusterHandlerFunc(a.patchClusterHandler)
	a.ClusterGetClusterHandler = cluster.GetClusterHandlerFunc(a.getClusterHandler)
}

/*
创建 cluster
1 如果有pvc，需要先创建pvc，需要注意pvc可以设置随deploy 或者不
2 创建deployemnt
3 创建 vm cr ，要提供projectid，token，deploy selflink, lb，floatingip等参数
*/
func (a *Api) createClusterHandler(params cluster.CreateClusterParams) middleware.Responder {
	return cluster.NewCreateClusterOK().WithPayload(nil)
}

/*
删除 cluster
1 若有随删除而删除pvc，则删除pvc
2 删除deployemnt
3 删除 vm cr
*/

func (a *Api) deleteClusterHandler(params cluster.DeleteClusterParams) middleware.Responder {
	return cluster.NewDeleteClusterOK().WithPayload(nil)
}

/*
更新 cluster
- label修改，替换+添加+删除
- 手动伸缩
- 扩容策略
*/
func (a *Api) patchClusterHandler(params cluster.PatchClusterParams) middleware.Responder {
	return cluster.NewPatchClusterOK().WithPayload(nil)
}

/*
查询 cluster
- 集群拓扑
- 各组件运行状态
- 规格
- 等等
*/
func (a *Api) getClusterHandler(params cluster.GetClusterParams) middleware.Responder {

	// containers := make([]corev1.Container, 1)
	// containers[0] = corev1.Container{
	// 	Name:  "container1",
	// 	Image: "nginx:latest",
	// }
	// creatPod := &corev1.Pod{
	// 	Spec: corev1.PodSpec{
	// 		Containers: containers,
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "nginx",
	// 		Namespace: "default",
	// 	},
	// }
	// err := a.ctrl.Create(creatPod)
	// if err != nil {
	// 	klog.Infof("create pod failed:%v", err)
	// 	return cluster.NewGetClusterOK().WithPayload(nil)
	// }

	retcode := int32(400)
	msg := "niubiplus"
	return cluster.NewGetClusterOK().WithPayload(&cluster.GetClusterOKBody{
		Code: &retcode,
		Data: &models.ClusterInfo{
			CloudeventsEndpoints: "xxx",
			GatewayEndpoints:     "zzz",
		},
		Message: &msg,
	})
	// return cluster.NewGetClusterOK().WithPayload(nil)
}
