/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	finopsv1 "github.com/krateoplatformops/finops-operator-vm-manager/api/v1"
)

// ConfigManagerVMReconciler reconciles a ConfigManagerVM object
type ConfigManagerVMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=finops.krateo.io,namespace=finops,resources=configmanagervms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=finops.krateo.io,namespace=finops,resources=configmanagervms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=finops.krateo.io,namespace=finops,resources=configmanagervms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigManagerVM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *ConfigManagerVMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Log.WithValues("FinOps.V1", req.NamespacedName)
	var err error

	var configManagerVM finopsv1.ConfigManagerVM
	if err = r.Get(ctx, req.NamespacedName, &configManagerVM); err != nil {
		logger.Info("unable to get current FocusConfig, probably deleted, ignoring...")
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}

	if configManagerVM.Spec.ProviderSpecificResources.AzureLogin.Action == "nop" {
		return ctrl.Result{}, nil
	}

	switch configManagerVM.Spec.ResourceProvider {
	case "azure":
		err = configManagerVM.Spec.ProviderSpecificResources.AzureLogin.Connect()
		if err != nil {
			return ctrl.Result{}, err
		}
		err = configManagerVM.Spec.ProviderSpecificResources.AzureLogin.SetResourceStatus()
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	configManagerVM.Spec.ProviderSpecificResources.AzureLogin.Action = "nop"
	err = r.Update(ctx, &configManagerVM)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigManagerVMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&finopsv1.ConfigManagerVM{}).
		Complete(r)
}
