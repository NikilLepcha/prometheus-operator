package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1alpha1 "github.com/NikilLepcha/prometheus-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrometheusOperatorReconciler reconciles a PrometheusOperator object
type PrometheusOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=monitoring.example.com,resources=prometheusoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.example.com,resources=prometheusoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.example.com,resources=prometheusoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PrometheusOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	prometheus := &monitoringv1alpha1.PrometheusOperator{}
	err := r.Get(ctx, req.NamespacedName, prometheus)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	dep := r.deploymentForPrometheus(prometheus)

	if err := controllerutil.SetControllerReference(prometheus, dep, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}
	
	size := prometheus.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PrometheusOperatorReconciler) deploymentForPrometheus(p *monitoringv1alpha1.PrometheusOperator) *appsv1.Deployment {
	labels := labelsForPrometheus(p.Name)
	replicas := p.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Name,
			Namespace: p.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: p.Spec.Image,
						Name:  "prometheus",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9090,
							Name:          "prometheus",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "prometheus-storage",
							MountPath: "/prometheus",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "prometheus-storage",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: fmt.Sprintf("%s-pvc", p.Name),
							},
						},
					}},
				},
			},
		},
	}

	return dep
}

func labelsForPrometheus(name string) map[string]string {
	return map[string]string {
		"app": "prometheus",
		"prometheus_cr": name,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrometheusOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.PrometheusOperator{}).
		Complete(r)
}
