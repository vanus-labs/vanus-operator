// Copyright 2017 EasyStack, Inc.
package handlers

import (
	"github.com/linkall-labs/vanus-operator/api/restapi/operations/connector"

	"github.com/go-openapi/runtime/middleware"
)

// var (
// 	WorkloadOldReplicAnno = "replic.ecns.io/workload"
// 	revisionAnnoKey       = "deployment.kubernetes.io/revision"
// 	rolloutReasonAnno     = "kubernetes.io/change-cause"
// )

// 所有注册的处理函数 应该 按照顺序出现在 Registxxx 下

func RegistConnectorHandler(a *Api) {
	a.ConnectorCreateConnectorHandler = connector.CreateConnectorHandlerFunc(a.createConnectorHandler)
	a.ConnectorDeleteConnectorHandler = connector.DeleteConnectorHandlerFunc(a.deleteConnectorHandler)
	a.ConnectorListConnectorHandler = connector.ListConnectorHandlerFunc(a.listConnectorHandler)
	a.ConnectorGetConnectorHandler = connector.GetConnectorHandlerFunc(a.getConnectorHandler)
}

/*
创建 conenctor
1 如果有pvc，需要先创建pvc，需要注意pvc可以设置随deploy 或者不
2 创建deployemnt
3 创建 vm cr ，要提供projectid，token，deploy selflink, lb，floatingip等参数
*/
func (a *Api) createConnectorHandler(params connector.CreateConnectorParams) middleware.Responder {
	return connector.NewCreateConnectorOK().WithPayload(nil)
}

/*
删除 connector
1 若有随删除而删除pvc，则删除pvc
2 删除deployemnt
3 删除 vm cr
*/

func (a *Api) deleteConnectorHandler(params connector.DeleteConnectorParams) middleware.Responder {
	return connector.NewDeleteConnectorOK().WithPayload(nil)
}

/*
列举 connector
- label修改，替换+添加+删除
- 手动伸缩
- 扩容策略
*/
func (a *Api) listConnectorHandler(params connector.ListConnectorParams) middleware.Responder {
	return connector.NewListConnectorOK().WithPayload(nil)
}

/*
查询 connector
- 集群拓扑
- 各组件运行状态
- 规格
- 等等
*/
func (a *Api) getConnectorHandler(params connector.GetConnectorParams) middleware.Responder {
	return connector.NewGetConnectorOK().WithPayload(nil)
}
