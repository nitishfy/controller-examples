package main

import (
	"context"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	podAnnotation = "configmap-name"
)

type PodConfigmapReconciler struct {
	client.Client
}

func main() {
	log := ctrl.Log.WithValues("configmap=examples")

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(manager).
	For(&corev1.Pod{}).
	Watches(
		&corev1.ConfigMap{},
		handler.EnqueueRequestsFromMapFunc(configMapToPodRequest(manager.GetClient())),
	).
	Complete(&PodConfigmapReconciler{Client: manager.GetClient()})
	if err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

// Reconcile syncs the state of a Pod based on the associated ConfigMap's data.
// It watches Pods and ConfigMaps. When a Pod with a specific annotation
// referencing a ConfigMap is created or updated, this method fetches the
// ConfigMap and updates the Pod's annotations with the key-value pairs from the
// ConfigMap's data field. If the ConfigMap has no data, the Pod's annotation
// "data" is set to "empty". Existing Pod annotations (other than those prefixed
// with "data.") are preserved. This ensures that changes in the referenced
// ConfigMap are reflected in the Pod annotations.
func (p *PodConfigmapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Log.WithValues("pod", req.NamespacedName)
	// get the pod from the Kubernetes API
	var pod corev1.Pod
	if err := p.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierror.IsNotFound(err) {
			logger.Error(err, "pod not found")
			return ctrl.Result{}, err
		}
	}

	// check if the pod has the annotation or not
	configMapName, exists := pod.Annotations[podAnnotation]
	if !exists {
		// do nothing if pod doesn't have the annotation
		return ctrl.Result{}, nil
	}

	// fetch the configmap and access the data field
	var configmap corev1.ConfigMap
	if err := p.Get(ctx, client.ObjectKey{
		Name: configMapName,
		Namespace: pod.Namespace,
	}, &configmap); err != nil {
		logger.Error(err, "error fetching the configmap")
		return ctrl.Result{}, nil
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	// access the data field in the configmap and add it to pod annotation
	if configmap.Data == nil || len(configmap.Data) == 0 {
		// clean up the old configmap data annotations (that are starting with data.)
		for key := range pod.Annotations {
			if strings.HasPrefix(key, "data.") {
				delete(pod.Annotations, key)
			}
		}
		pod.Annotations["data"] = "empty"
	} else {
		for k, v := range configmap.Data {
			pod.Annotations["data."+k] = v
		}
	}
	
	// update the pod spec
	if err := p.Update(ctx, &pod); err != nil {
		logger.Error(err, "failed to update pod with configmap data")
		return ctrl.Result{}, err
	}

	logger.Info("successfully updated pod annotations from configmap")
	return ctrl.Result{}, nil 
}

// configMapToPodRequest returns a handler.MapFunc that maps ConfigMap events to
// reconcile.Requests for Pods. When a ConfigMap changes, this function lists
// all Pods in the same namespace and filters those with an annotation
// referencing this ConfigMap's name. It then creates reconcile.Requests for
// those Pods to trigger their reconciliation.
func configMapToPodRequest(c client.Client) handler.MapFunc {
	return func (ctx context.Context, obj client.Object) []reconcile.Request {
		cm := obj.(*corev1.ConfigMap)
		log := log.Log.WithValues("configamp", cm.Namespace)
		
		var pods corev1.PodList
		if err := c.List(ctx, &pods, client.InNamespace(cm.Namespace)); err != nil {
			log.Error(err, "error listing pods")
			return nil
		}

		var requests []reconcile.Request
		for _, pod := range pods.Items {
			if pod.Annotations[podAnnotation] == cm.Name {
				requests = append(requests, reconcile.Request{
					NamespacedName: client.ObjectKey {
						Name: pod.Name,
						Namespace: pod.Namespace,
					},
				})
			}
		}

		return requests
	}
}
