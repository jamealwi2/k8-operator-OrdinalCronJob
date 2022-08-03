/*
Copyright 2022.

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
	"reflect"
	"strconv"

	batchV1 "k8s.io/api/batch/v1"
	batchV1beta1 "k8s.io/api/batch/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	scheduledjobsv1 "github.com/jamealwi2/ordinalcronjob/api/v1"
)

// OrdinalCronJobReconciler reconciles a OrdinalCronJob object
type OrdinalCronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=scheduledjobs.jamealwi,resources=ordinalcronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scheduledjobs.jamealwi,resources=ordinalcronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scheduledjobs.jamealwi,resources=ordinalcronjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OrdinalCronJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *OrdinalCronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconciling ordinalCronJob custom resource")

	var ordinalCronJob scheduledjobsv1.OrdinalCronJob
	if err := r.Get(ctx, req.NamespacedName, &ordinalCronJob); err != nil {
		log.Error(err, "unable to fetch ordinalCronJob")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ordinalCronJob resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info(ordinalCronJob.Name + "|" + ordinalCronJob.Spec.CronSchedule + "|" + ordinalCronJob.Spec.JobsRequired)

	nJobs, err := strconv.Atoi(ordinalCronJob.Spec.JobsRequired)
	if err != nil {
		log.Error(err, "Jobs reequired should be integer")
	}

	listCj := &batchV1beta1.CronJobList{}
	r.List(context.TODO(), listCj, &client.ListOptions{Namespace: ordinalCronJob.Namespace, Raw: &metav1.ListOptions{
		FieldSelector: "metadata.ownerReferences[0].uid=" + string(ordinalCronJob.UID),
	}})

	i := 1
	for ; i <= nJobs; i++ {
		cjName := ordinalCronJob.Name + "-cj-" + strconv.Itoa(i)
		cj1 := &batchV1beta1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cjName,
				Namespace: ordinalCronJob.Namespace,
				Labels: map[string]string{
					"app": ordinalCronJob.Name,
					"id":  strconv.Itoa(i),
				},
				Annotations: ordinalCronJob.Annotations,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: ordinalCronJob.APIVersion,
						Kind:       ordinalCronJob.Kind,
						Name:       ordinalCronJob.Name,
						UID:        ordinalCronJob.UID,
					},
				},
			},
			Spec: batchV1beta1.CronJobSpec{
				Schedule:                ordinalCronJob.Spec.CronSchedule,
				Suspend:                 ordinalCronJob.Spec.Suspend,
				StartingDeadlineSeconds: ordinalCronJob.Spec.StartingDeadlineSeconds,
				JobTemplate: batchV1beta1.JobTemplateSpec{
					Spec: batchV1.JobSpec{
						Template: apiv1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app": ordinalCronJob.Name,
									"id":  strconv.Itoa(i),
								},
								Annotations: ordinalCronJob.Annotations,
							},
							Spec: apiv1.PodSpec{
								RestartPolicy:    "Never",
								ImagePullSecrets: ordinalCronJob.Spec.ImagePullSecrets,
								Containers: []apiv1.Container{
									{
										Name:         ordinalCronJob.Name,
										Image:        ordinalCronJob.Spec.Image,
										Command:      ordinalCronJob.Spec.Command,
										Args:         ordinalCronJob.Spec.Args,
										Resources:    ordinalCronJob.Spec.Resources,
										VolumeMounts: ordinalCronJob.Spec.VolumeMounts,
										Env:          ordinalCronJob.Spec.Env,
									},
								},
								Volumes: ordinalCronJob.Spec.Volumes,
							},
						},
					},
				},
			},
		}

		found := &batchV1beta1.CronJob{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: cjName, Namespace: ordinalCronJob.Namespace}, found)
		if err == nil && found.Name != "" {
			if !reflect.DeepEqual(cj1.Spec, found.Spec) {
				log.Info("updating cronjob...")
				err := r.Update(context.TODO(), cj1)
				if err != nil {
					panic(err)
				}
			}
		} else {
			log.Info("creating cronjob...")
			err := r.Create(context.TODO(), cj1)
			if err != nil {
				panic(err)
			}
		}

		ctrl.SetControllerReference(&ordinalCronJob, cj1, r.Scheme)
	}

	for ; i <= len(listCj.Items); i++ {
		cjName := ordinalCronJob.Name + "-cj-" + strconv.Itoa(i)
		log.Info("deleting..." + cjName)
		cj1 := &batchV1beta1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cjName,
				Namespace: ordinalCronJob.Namespace,
			},
		}
		r.Delete(context.TODO(), cj1)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrdinalCronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scheduledjobsv1.OrdinalCronJob{}).
		Complete(r)
}
