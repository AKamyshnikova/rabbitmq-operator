package rabbitmq

import (
	rabbitmqv1alpha1 "github.com/toha10/rabbitmq-operator/pkg/apis/rabbitmq/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newService(cr *rabbitmqv1alpha1.RabbitMQ) *corev1.Service {
	labels := map[string]string{
		"app":  "rabbitmq",
		"type": "LoadBalancer",
	}
	selector := map[string]string{"app": "rabbitmq"}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.DiscoveryService,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     15672,
				},
				{
					Name:     "amqp",
					Protocol: corev1.ProtocolTCP,
					Port:     5672,
				},
			},
		},
	}
}

func newConfigMap(cr *rabbitmqv1alpha1.RabbitMQ) *corev1.ConfigMap {
	rabbitmqPlugins := "[rabbitmq_management,rabbitmq_peer_discovery_k8s]."

	rabbitmqConf := `## Cluster formation. See https://www.rabbitmq.com/cluster-formation.html to learn more.
cluster_formation.peer_discovery_backend  = rabbit_peer_discovery_k8s
cluster_formation.k8s.host = kubernetes.default.svc
## Should RabbitMQ node name be computed from the pod's hostname or IP address?
## IP addresses are not stable, so using [stable] hostnames is recommended when possible.
## Set to "hostname" to use pod hostnames.
## When this value is changed, so should the variable used to set the RABBITMQ_NODENAME
## environment variable.
cluster_formation.k8s.address_type = ip
## How often should node cleanup checks run?
cluster_formation.node_cleanup.interval = 30
## Set to false if automatic removal of unknown/absent nodes
## is desired. This can be dangerous, see
##  * https://www.rabbitmq.com/cluster-formation.html#node-health-checks-and-cleanup
##  * https://groups.google.com/forum/#!msg/rabbitmq-users/wuOfzEywHXo/k8z_HWIkBgAJ
cluster_formation.node_cleanup.only_log_warning = false
cluster_partition_handling = autoheal
## See https://www.rabbitmq.com/ha.html#master-migration-data-locality
queue_master_locator=min-masters
## See https://www.rabbitmq.com/access-control.html#loopback-users
loopback_users.guest = false`

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-config",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"rabbitmq.conf":   rabbitmqConf,
			"enabled_plugins": rabbitmqPlugins,
		},
	}
}

func newStatefulSet(cr *rabbitmqv1alpha1.RabbitMQ) *v1.StatefulSet {
	labels := map[string]string{
		"app": "rabbitmq",
	}
	dataVolumeName := "rabbitmq-data"
	podContainers := []corev1.Container{}

	// container with rabbitmq
	rabbitmqContainer := corev1.Container{
		Name:  "rabbitmq",
		Image: cr.Spec.Image,
		Env: []corev1.EnvVar{
			{
				Name: "MY_POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
			{
				Name:  "RABBITMQ_USE_LONGNAME",
				Value: "true",
			},
			{
				Name:  "RABBITMQ_NODENAME",
				Value: "rabbit@$(MY_POD_IP)",
			},
			{
				Name:  "K8S_SERVICE_NAME",
				Value: cr.Spec.DiscoveryService,
			},
			{
				Name:  "RABBITMQ_ERLANG_COOKIE",
				Value: "mycookie",
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: 15672,
			},
			{
				Name:          "amqp",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: 5672,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config-volume",
				MountPath: "/etc/rabbitmq",
			},
			{
				Name:      dataVolumeName,
				MountPath: "/var/lib/rabbitmq",
			},
		},
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: 20,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"rabbitmqctl",
						"status",
					},
				},
			},
		},
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: 60,
			TimeoutSeconds:      15,
			PeriodSeconds:       60,
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"rabbitmqctl",
						"status",
					},
				},
			},
		},
	}

	if cr.Spec.Affinity == nil {
		cr.Spec.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 20,
						PodAffinityTerm: corev1.PodAffinityTerm{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "app",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{cr.GetName()},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	if cr.Spec.Resources != nil {
		rabbitmqContainer.Resources = *cr.Spec.Resources
	}

	//  RABBITMQ_DEFAULT_USER
	if len(cr.Spec.DefaultUsername) > 0 {
		rabbitmqContainer.Env = append(rabbitmqContainer.Env, corev1.EnvVar{
			Name:  "RABBITMQ_DEFAULT_USER",
			Value: cr.Spec.DefaultUsername,
		})
	}

	//  RABBITMQ_DEFAULT_PASS
	if len(cr.Spec.DefaultPassword) > 0 {
		rabbitmqContainer.Env = append(rabbitmqContainer.Env, corev1.EnvVar{
			Name:  "RABBITMQ_DEFAULT_PASS",
			Value: cr.Spec.DefaultPassword,
		})
	}

	// RABBITMQ_DEFAULT_VHOST
	if len(cr.Spec.DefaultVHost) > 0 {
		rabbitmqContainer.Env = append(rabbitmqContainer.Env, corev1.EnvVar{
			Name:  "RABBITMQ_DEFAULT_VHOST",
			Value: cr.Spec.DefaultVHost,
		})
	}

	podContainers = append(podContainers, rabbitmqContainer)

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: cr.Spec.ServiceAccount,
			Containers:         podContainers,
			Volumes: []corev1.Volume{
				{
					Name: "config-volume",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "rabbitmq-config",
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "rabbitmq.conf",
									Path: "rabbitmq.conf",
								},
								{
									Key:  "enabled_plugins",
									Path: "enabled_plugins",
								},
							},
						},
					},
				},
			},
			Affinity: cr.Spec.Affinity,
		},
	}

	ss := &v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: v1.StatefulSetSpec{
			Replicas:    &cr.Spec.Replicas,
			Template:    podTemplate,
			ServiceName: cr.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: v1.StatefulSetUpdateStrategy{
				Type: v1.OnDeleteStatefulSetStrategyType,
			},
		},
	}

	if !cr.Spec.DataVolumeSize.IsZero() {
		pvcTemplates := []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rabbitmq-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},

					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: cr.Spec.DataVolumeSize,
						},
					},

					StorageClassName: &cr.Spec.DataStorageClass,
				},
			},
		}

		ss.Spec.VolumeClaimTemplates = pvcTemplates
	} else {
		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, corev1.Volume{
			Name:         dataVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}

	return ss
}
