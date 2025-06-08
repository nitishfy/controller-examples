package main

import (
	"context"
	"os"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	nsLabelKey = "secret-sync"
)

type SecretReconciler struct {
	client.Client
}

func main() {
	log := ctrl.Log.WithName("secret-synchronizer")

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(manager).
		For(&corev1.Secret{}).
		Complete(&SecretReconciler{Client: manager.GetClient()})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

func (s *SecretReconciler) getCentralSecret(ctx context.Context, logger logr.Logger) (*corev1.Secret, error) {
	var centralSecret corev1.Secret
	if err := s.Get(ctx, client.ObjectKey {
		Namespace: "default",
		Name: "central-secret",
	}, &centralSecret); err != nil {
		if apierror.IsNotFound(err) {
			logger.Error(err, "central secret not found in the default namespace")
			return nil, nil
		}
		return nil, err
	}

	return &centralSecret, nil
}

func (s *SecretReconciler) listNamespaces(ctx context.Context, logger logr.Logger) ([]corev1.Namespace, error) {
	var nsList corev1.NamespaceList
	if err := s.List(ctx, &nsList); err != nil {
		logger.Error(err, "error listing the namespaces")
		return nil, err
	}

	var result []corev1.Namespace
	for _, ns := range nsList.Items {
		// check if the namespace has the label key 'secret-sync' set to "true"
		// and add it to the result
		if val, ok := ns.Labels[nsLabelKey]; ok && val == "true" {
			result = append(result, ns)
		}
	}

	return result, nil
}

func (s *SecretReconciler) createOrUpdateSecretInNamespace(ctx context.Context, ns corev1.Namespace, centralSecret *corev1.Secret, logger logr.Logger) error {
		var existingSecret corev1.Secret
		desiredSecret := &corev1.Secret {
			ObjectMeta: v1.ObjectMeta{
				Name: "central-secret",
				Namespace: ns.Name,
			},
			Data: centralSecret.Data,
			Type: centralSecret.Type,
		}
		
		err := s.Get(ctx, client.ObjectKey{
			Namespace: ns.Name,
			Name: "central-secret",
		}, &existingSecret)

		switch {
		// if the descired secret is not found, create it
		case apierror.IsNotFound(err):
			if err := s.Create(ctx, desiredSecret); err != nil {
				logger.Error(err, "failed to create secret", "namespace", ns.Name)
				return err
			}
		// if secret exists, check for relevant fields and update it if required	
		case err == nil:
			if !reflect.DeepEqual(existingSecret.Data, centralSecret.Data) {
				existingSecret.Data = centralSecret.Data
				existingSecret.Type = centralSecret.Type
				if err := s.Update(ctx, &existingSecret); err != nil {
					logger.Error(err, "failed to update the secret in namespace", ns.Namespace)
					return err
				}
				logger.Info("updated secret", "namespace", ns.Name)
			}
		default:
			logger.Error(err, "failed to get secret", "namespace", ns.Name)
			return err	
		}

		return nil
}

func (s *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Log.WithValues("secret", req.NamespacedName)

	// fetch the central secret from the default namespace
	centralSecret, err := s.getCentralSecret(ctx, logger)
	if err != nil || centralSecret == nil {
		// TODO(@nitishfy) handle the case when central secret gets deleted 
		// in that case, delete the central-secret from other namespaces too
		return ctrl.Result{}, err
	}

	// filter the namespaces with the label set
	namespaces, err := s.listNamespaces(ctx, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	// iterate over the filtered namespaces and create/update secret in those namespace
	for _, ns := range namespaces {
		if err := s.createOrUpdateSecretInNamespace(ctx, ns, centralSecret, logger); err != nil {
			logger.Error(err, "failed to syn secret", "namespace", ns.Name)
		}
	}

	return ctrl.Result{}, nil
}
