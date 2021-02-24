/*
Copyright 2020 The Terrak8s-operator authors.

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

package controllers

import (
	"context"
	"fmt"
	sqlv1alpha1 "github.com/HamzaZo/terrak8s-operator/api/v1alpha1"
	"github.com/HamzaZo/terrak8s-operator/pkg/terraform"
	"github.com/HamzaZo/terrak8s-operator/pkg/util"
	"github.com/go-logr/logr"
	"io/ioutil"
	kubeApiV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"
)

const (

	// SuccessSynced is used as part of the Event 'reason' when Store resource is synced
	SuccessSynced = "Synced"
	// MessageResourceSynced is the message used for an Event fired when Store resource
	// is synced successfully
	MessageResourceSynced = "PostgreSql Resource synced successfully"
	//Finalizer name of  finalizer
	Finalizer = "sql.terrak8s.io"
)

var (
	secretList kubeApiV1.SecretList
	dir        string
)

// PostgreSqlReconciler reconciles a PostgreSql object
type PostgreSqlReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=sql.terrak8s.io,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sql.terrak8s.io,resources=postgresqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *PostgreSqlReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("postgresql", req.NamespacedName)

	// Fetch postgresql instance
	instance := &sqlv1alpha1.PostgreSql{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request - return and don't requeue:
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch PostgreSql")
		return ctrl.Result{}, err
	}

	if util.IsBeingDeleted(instance) {
		if instance.Status.Phase == sqlv1alpha1.PhaseFailed {
			util.RemoveFinalizer(instance, Finalizer)
			if errU := r.Update(context.Background(), instance); errU != nil {
				return ctrl.Result{}, errU
			}
			return ctrl.Result{}, nil
		} else {
			errD := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseDestroying)
			if errD != nil {
				return ctrl.Result{}, errD
			}
			errs := terraform.Destroy(dir)
			if errs != nil {
				errMsg := fmt.Sprintf("failed to destroy instance %v/%v ", instance.Namespace, instance.Name)
				log.Error(errs, errMsg)
				return ctrl.Result{Requeue: true}, nil
			}
			errC := util.HouseCleaning(dir)
			if errC != nil {
				errMsg := fmt.Sprintf("failed to do houseCleaning for instance %v/%v ", instance.Namespace, instance.Name)
				r.Log.Error(errC, errMsg)
				return ctrl.Result{Requeue: true}, nil
			}
			util.RemoveFinalizer(instance, Finalizer)
			if errU := r.Update(context.Background(), instance); errU != nil {
				return ctrl.Result{}, errU
			}
			// Stop reconciliation as the item is being deleted
			return ctrl.Result{}, nil
		}
	} else {
		util.AddFinalizer(instance, Finalizer)
		if err := r.Update(context.Background(), instance); err != nil {
			return ctrl.Result{}, err
		}
		dir, err = util.CreateDirectory(instance.Namespace, instance.Name)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create instance %v/%v tf dir", instance.Namespace, instance.Name)
			r.Log.Error(err, errMsg)
			return ctrl.Result{}, err
		}

	}

	b, err := r.FetchUserPasswordFromSecret(req.Namespace, instance, ctx, secretList)
	if err != nil {
		errUp := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseFailed)
		if errUp != nil {
			return ctrl.Result{}, errUp
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	errS := r.GetGCPCredentialsFromSecret(secretList, req.Namespace, ctx, instance, dir)
	if errS != nil {
		errUp := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseFailed)
		if errUp != nil {
			return ctrl.Result{}, errUp
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	errF := r.GenerateTFFromCR(instance, dir, b)
	if errF != nil {
		return ctrl.Result{}, errF
	}

	errUp := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseInitializing)
	if errUp != nil {
		return ctrl.Result{}, err
	}
	errSo := r.ProvisioningStorageBucket(dir, instance, ctx)
	if errSo != nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}
	r.Recorder.Eventf(instance, kubeApiV1.EventTypeNormal, "SuccessfulApplying", "successfully provision storage bucket %q", instance.Spec.RemoteState.BucketName)

	errI := r.InitializeRemoteBackend(dir, instance, ctx)
	if errI != nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	r.Recorder.Eventf(instance, kubeApiV1.EventTypeNormal, "SuccessfulInitialize", "successfully configured the remote backend \"gcs\" bucket")

	errAp := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseApplying)
	if errAp != nil {
		return ctrl.Result{}, err
	}
	errP := r.ProvisioningInstance(dir, instance, ctx)
	if errP != nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}
	r.Recorder.Eventf(instance, kubeApiV1.EventTypeNormal, "SuccessfullyApplying", "successfully creating cloud sql instance %q", instance.Name)

	out, errO := r.GetOutput(dir, instance)
	if errO != nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}
	errR := r.UpdateStatus(ctx, out, sqlv1alpha1.PhaseRunning)
	if errR != nil {
		return ctrl.Result{}, err
	}

	log.Info("resource status synced")

	r.Recorder.Event(instance, kubeApiV1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	// Don't requeue. We should be reconcile because the CR changes.
	return ctrl.Result{}, nil
}

func (r *PostgreSqlReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&sqlv1alpha1.PostgreSql{}).
		WithEventFilter(pred).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Complete(r)
}

//GetGCPCredentialsFromSecret fetch gcp serviceAccount from secret
func (r *PostgreSqlReconciler) GetGCPCredentialsFromSecret(secretList kubeApiV1.SecretList, namespace string, ctx context.Context, instance *sqlv1alpha1.PostgreSql, dir string) error {
	var filePath string
	err := r.List(ctx, &secretList, client.InNamespace(namespace))
	if err != nil {
		errMsg := fmt.Sprintf("unable to list secret in namespace  %v", instance.Namespace)
		r.Log.Error(err, errMsg)
	}
	isFound := false
	for _, k := range secretList.Items {
		for obj := range k.Data {
			if !strings.Contains(obj, ".json") {
				continue
			}
			isFound = true
			value := k.Data[obj]
			filePath = filepath.Join(dir + "/" + obj)
			if len(value) != 0 {
				if err := ioutil.WriteFile(filePath, value, 0600); err != nil {
					return fmt.Errorf("failed to write serviceAccount json key to file %v", err)
				}
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filePath)
			}
		}
	}
	if !isFound {
		r.Recorder.Eventf(instance, kubeApiV1.EventTypeWarning, "KeyNotFound", "unable to find valid GCP serviceAccount json key in namespace %q", namespace)
		return fmt.Errorf("json secret key not found in namespace %v", namespace)
	}
	return nil
}

//FetchUserPasswordFromSecret fetch secret from namespace based on CR
func (r *PostgreSqlReconciler) FetchUserPasswordFromSecret(namespace string, instance *sqlv1alpha1.PostgreSql, ctx context.Context, secretList kubeApiV1.SecretList) (map[string][]byte, error) {
	secretCred := make(map[string][]byte)
	err := r.List(ctx, &secretList, client.InNamespace(namespace))
	if err != nil {
		errMsg := fmt.Sprintf("unable to list secret in namespace %v", instance.Namespace)
		r.Log.Error(err, errMsg)
		return nil, err
	}
	for i, j := range GetSecretFromCR(instance) {
		isFound := false
		for _, k := range secretList.Items {
			if strings.TrimSpace(i) != k.Name {
				continue
			}
			isFound = true
			for _, p := range j {
				if _, exists := k.Data[p]; !exists {
					r.Recorder.Eventf(instance, kubeApiV1.EventTypeWarning, "KeyNotFound", "unable to find secret key %q", p)
					return nil, fmt.Errorf("secret key %q/%q does not exist %v", p, namespace, err)
				}
				if util.IsValidPasswordFormat(string(k.Data[p])) {
					secretCred[p] = k.Data[p]
				} else {
					r.Recorder.Eventf(instance, kubeApiV1.EventTypeWarning, "InvalidPassword", "secret key %q. Must respect password rules: at least 7 letters - at least 1 number - at least 1 upper case - at least 1 special character", p)
					return nil, fmt.Errorf("secret %q/%q does not respect password rules - error %v", p, namespace, err)
				}

			}

		}
		if !isFound {
			r.Recorder.Eventf(instance, kubeApiV1.EventTypeWarning, "KeyNotFound", "unable to find secret %q", i)
			return nil, fmt.Errorf("secret %v does not exist in namespace %v - error %v", i, namespace, err)
		}
	}
	return secretCred, nil
}

//UpdateStatus Update the CR status
func (r *PostgreSqlReconciler) UpdateStatus(ctx context.Context, instance *sqlv1alpha1.PostgreSql, phase sqlv1alpha1.ObjectPhase) error {
	instance.Status.Phase = phase
	err := r.Status().Update(ctx, instance)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update PostgreSql instance status  %v/%v", instance.Name, instance.Namespace)
		r.Log.Error(err, errMsg)
		return err
	}
	return nil
}

//GenerateTFFromCR generate tf files from CR
func (r *PostgreSqlReconciler) GenerateTFFromCR(instance *sqlv1alpha1.PostgreSql, dir string, value map[string][]byte) error {
	errMsg := fmt.Sprintf("failed to generate tf files  %v/%v", instance.Name, instance.Namespace)
	errB := terraform.GenerateBucketTF(instance, filepath.Join(dir, "bucket"))
	if errB != nil {
		r.Log.Error(errB, errMsg)
		return errB
	}
	errP := terraform.GenerateProviderAndBackendTF(instance, filepath.Join(dir, "instance"))
	if errP != nil {
		r.Log.Error(errP, errMsg)
		return errP
	}
	errI := terraform.GenerateTFInstance(instance, filepath.Join(dir, "instance"), value)
	if errI != nil {
		r.Log.Error(errI, errMsg)
		return errI
	}
	errO := terraform.GenerateTFOutput(filepath.Join(dir, "instance"))
	if errO != nil {
		r.Log.Error(errO, errMsg)
		return errO
	}
	return nil
}

//ProvisioningStorageBucket provision storage bucket based on generated tf
func (r *PostgreSqlReconciler) ProvisioningStorageBucket(dir string, bucket *sqlv1alpha1.PostgreSql, ctx context.Context) error {
	err := terraform.Init(filepath.Join(dir, "bucket"))
	if err != nil {
		initMsg := fmt.Sprintf("initializing storage bucket failed %v", bucket.Spec.BucketConfig.Name)
		r.Log.Error(err, initMsg)

		errUp := r.UpdateStatus(ctx, bucket, sqlv1alpha1.PhaseFailed)
		r.Recorder.Eventf(bucket, kubeApiV1.EventTypeWarning, "InitializeFailed ", "failed to initialize storage bucket %q", bucket.Spec.BucketConfig.Name)
		if errUp != nil {
			return errUp
		}
		return err
	}
	err = terraform.Apply(filepath.Join(dir, "bucket"))
	if err != nil {
		applyMsg := fmt.Sprintf("provisioning storage bucket failed %v", bucket.Spec.BucketConfig.Name)
		r.Log.Error(err, applyMsg)
		errUp := r.UpdateStatus(ctx, bucket, sqlv1alpha1.PhaseFailed)

		r.Recorder.Eventf(bucket, kubeApiV1.EventTypeWarning, "ApplyingFailed ", "failed to provision storage bucket %q - bucket names must be unique", bucket.Spec.BucketConfig.Name)
		if errUp != nil {
			return errUp
		}
		return err
	}
	return nil
}

//InitializeRemoteBackend initialize remote backend based on generated tf
func (r *PostgreSqlReconciler) InitializeRemoteBackend(dir string, instance *sqlv1alpha1.PostgreSql, ctx context.Context) error {
	err := terraform.Init(filepath.Join(dir, "instance"))
	if err != nil {
		errMsg := fmt.Sprintf("initializing remote backend failed for instance %v/%v", instance.Name, instance.Namespace)
		r.Log.Error(err, errMsg)

		errUp := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseFailed)
		r.Recorder.Eventf(instance, kubeApiV1.EventTypeWarning, "InitializeFailed ", "failed to initialize remote backend %q", instance.Spec.RemoteState.BucketName)
		if errUp != nil {
			return errUp
		}
		return err
	}
	return nil
}

//ProvisioningInstance provision sql instance based on generated tf
func (r *PostgreSqlReconciler) ProvisioningInstance(dir string, instance *sqlv1alpha1.PostgreSql, ctx context.Context) error {
	err := terraform.Apply(filepath.Join(dir, "instance"))
	if err != nil {
		errMsg := fmt.Sprintf("provisioning sql instance  %v/%v failed", instance.Name, instance.Namespace)
		r.Log.Error(err, errMsg)

		errUp := r.UpdateStatus(ctx, instance, sqlv1alpha1.PhaseFailed)
		r.Recorder.Eventf(instance, kubeApiV1.EventTypeWarning, "ApplyingFailed ", "failed to provision cloud sql instance %q", instance.Name)
		if errUp != nil {
			return errUp
		}
		return err
	}
	return nil
}

//GetOutput get output and update the output status
func (r *PostgreSqlReconciler) GetOutput(dir string, instance *sqlv1alpha1.PostgreSql) (*sqlv1alpha1.PostgreSql, error) {
	output, errO := terraform.Output(filepath.Join(dir, "instance"))
	if errO != nil {
		errMsg := fmt.Sprintf("failed to get instance %v/%v output ", instance.Name, instance.Namespace)
		r.Log.Error(errO, errMsg)
		return nil, errO
	}
	errs, out := util.UpdateOutput(instance, output)
	if errs != nil {
		errMsg := fmt.Sprintf("failed to update instance %v/%v ", instance.Name, instance.Namespace)
		r.Log.Error(errs, errMsg)
		return nil, errs
	}
	return out, nil
}

//GetSecretFromCR stores secretKeyRefs on a map
func GetSecretFromCR(instance *sqlv1alpha1.PostgreSql) map[string][]string {
	secrets := make(map[string][]string)
	var names []string
	var keys []string
	for _, k := range instance.Spec.Users {
		names = append(names, k.Password.SecretKeyRef.Name)
		keys = append(keys, k.Password.SecretKeyRef.Key)
		for _, i := range names {
			secrets[i] = keys
		}
	}
	return secrets
}
