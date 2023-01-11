package handlers

import (
	"context"
	"encoding/json"

	vanusv1alpha1 "github.com/linkall-labs/vanus-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	ResourceController = "controllers"
	ResourceStore      = "stores"
	ResourceTrigger    = "triggers"
	ResourceTimer      = "timers"
	ResourceGateway    = "gateways"
)

func (a *Api) createController(controller *vanusv1alpha1.Controller, namespace string) (*vanusv1alpha1.Controller, error) {
	existController, exist, err := a.existController(namespace, controller.Name, &metav1.GetOptions{})
	if err != nil {
		return existController, err
	}
	if exist {
		return existController, nil
	}
	result := &vanusv1alpha1.Controller{}
	err = a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceController).
		Body(controller).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchController(controller *vanusv1alpha1.Controller) (*vanusv1alpha1.Controller, error) {
	result := &vanusv1alpha1.Controller{}
	body, err := json.Marshal(controller)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(controller.Namespace).
		Name(controller.Name).
		Resource(ResourceController).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteController(namespace string, name string) error {
	_, exist, err := a.existController(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Controller{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceController).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listController(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.ControllerList, error) {
	result := &vanusv1alpha1.ControllerList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceController).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getController(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Controller, error) {
	result := &vanusv1alpha1.Controller{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceController).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existController(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Controller, bool, error) {
	result := &vanusv1alpha1.Controller{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceController).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}

func (a *Api) createStore(store *vanusv1alpha1.Store, namespace string) (*vanusv1alpha1.Store, error) {
	result := &vanusv1alpha1.Store{}
	err := a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceStore).
		Body(store).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchStore(store *vanusv1alpha1.Store) (*vanusv1alpha1.Store, error) {
	result := &vanusv1alpha1.Store{}
	body, err := json.Marshal(store)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(store.Namespace).
		Name(store.Name).
		Resource(ResourceStore).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteStore(namespace string, name string) error {
	_, exist, err := a.existStore(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Store{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceStore).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listStore(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.StoreList, error) {
	result := &vanusv1alpha1.StoreList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceStore).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getStore(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Store, error) {
	result := &vanusv1alpha1.Store{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceStore).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existStore(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Store, bool, error) {
	result := &vanusv1alpha1.Store{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceStore).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}

func (a *Api) createTrigger(trigger *vanusv1alpha1.Trigger, namespace string) (*vanusv1alpha1.Trigger, error) {
	result := &vanusv1alpha1.Trigger{}
	err := a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceTrigger).
		Body(trigger).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchTrigger(trigger *vanusv1alpha1.Trigger) (*vanusv1alpha1.Trigger, error) {
	result := &vanusv1alpha1.Trigger{}
	body, err := json.Marshal(trigger)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(trigger.Namespace).
		Name(trigger.Name).
		Resource(ResourceTrigger).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteTrigger(namespace string, name string) error {
	_, exist, err := a.existTrigger(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Trigger{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceTrigger).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listTrigger(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.TriggerList, error) {
	result := &vanusv1alpha1.TriggerList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceTrigger).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getTrigger(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Trigger, error) {
	result := &vanusv1alpha1.Trigger{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceTrigger).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existTrigger(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Trigger, bool, error) {
	result := &vanusv1alpha1.Trigger{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceTrigger).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}

func (a *Api) createTimer(timer *vanusv1alpha1.Timer, namespace string) (*vanusv1alpha1.Timer, error) {
	result := &vanusv1alpha1.Timer{}
	err := a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceTimer).
		Body(timer).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchTimer(timer *vanusv1alpha1.Timer) (*vanusv1alpha1.Timer, error) {
	result := &vanusv1alpha1.Timer{}
	body, err := json.Marshal(timer)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(timer.Namespace).
		Name(timer.Name).
		Resource(ResourceTimer).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteTimer(namespace string, name string) error {
	_, exist, err := a.existTimer(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Timer{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceTimer).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listTimer(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.TimerList, error) {
	result := &vanusv1alpha1.TimerList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceTimer).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getTimer(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Timer, error) {
	result := &vanusv1alpha1.Timer{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceTimer).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existTimer(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Timer, bool, error) {
	result := &vanusv1alpha1.Timer{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceTimer).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}

func (a *Api) createGateway(store *vanusv1alpha1.Gateway, namespace string) (*vanusv1alpha1.Gateway, error) {
	result := &vanusv1alpha1.Gateway{}
	err := a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceGateway).
		Body(store).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchGateway(gateway *vanusv1alpha1.Gateway) (*vanusv1alpha1.Gateway, error) {
	result := &vanusv1alpha1.Gateway{}
	body, err := json.Marshal(gateway)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(gateway.Namespace).
		Name(gateway.Name).
		Resource(ResourceGateway).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteGateway(namespace string, name string) error {
	_, exist, err := a.existGateway(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Gateway{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceGateway).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listGateway(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.GatewayList, error) {
	result := &vanusv1alpha1.GatewayList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceGateway).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getGateway(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Gateway, error) {
	result := &vanusv1alpha1.Gateway{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceGateway).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existGateway(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Gateway, bool, error) {
	result := &vanusv1alpha1.Gateway{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceGateway).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}
