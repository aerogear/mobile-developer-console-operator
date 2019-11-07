package mobiledeveloperconsole

import (
	"fmt"
	"github.com/aerogear/mobile-developer-console-operator/pkg/constants"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/aerogear/mobile-developer-console-operator/pkg/util"
	"github.com/aerogear/mobile-developer-console-operator/version"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	mdcv1alpha1 "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newMDCServiceAccount(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.ServiceAccount, error) {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.mdc": fmt.Sprintf("{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"%s-mdc-proxy\"}}", cr.Name),
			},
		},
	}, nil
}

func newOauthProxyService(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: util.ObjectMeta(&cr.ObjectMeta, "mdc-proxy"),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "mdc",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4180,
					},
				},
			},
		},
	}, nil
}

func newMDCService(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.Service, error) {
	serviceObjectMeta := util.ObjectMeta(&cr.ObjectMeta, "mdc")
	serviceObjectMeta.Labels["mobile"] = "enabled"
	serviceObjectMeta.Labels["internal"] = "mdc"

	return &corev1.Service{
		ObjectMeta: serviceObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "mdc",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4000,
					},
				},
			},
		},
	}, nil
}

func newOauthProxyRoute(cr *mdcv1alpha1.MobileDeveloperConsole) (*routev1.Route, error) {
	return &routev1.Route{
		ObjectMeta: util.ObjectMeta(&cr.ObjectMeta, "mdc-proxy"),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("%s-%s", cr.Name, "mdc-proxy"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
			},
		},
	}, nil
}

func newMDCDeployment(cr *mdcv1alpha1.MobileDeveloperConsole) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app":     cr.Name,
		"service": "mdc",
	}

	cookieSecret, err := util.GeneratePassword()
	if err != nil {
		return nil, errors.Wrap(err, "error generating cookie secret")
	}

	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
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
					ServiceAccountName: cr.Name,
					Containers: []corev1.Container{
						{
							Name:            cfg.MDCContainerName,
							Image:           constants.MDCImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: cr.Namespace,
								},
								{
									Name:  "NODE_ENV",
									Value: "production",
								},
								{
									Name:  "OPENSHIFT_HOST",
									Value: cfg.OpenShiftHost,
								},
								{
									Name:  "IDM_DOCUMENTATION_URL",
									Value: cfg.IdentityManagementDocumentationURL,
								},
								{
									Name:  "UPS_DOCUMENTATION_URL",
									Value: cfg.UnifiedPushDocumentationURL,
								},
								{
									Name:  "SYNC_DOCUMENTATION_URL",
									Value: cfg.DataSyncDocumentationURL,
								},
								{
									Name:  "MSS_DOCUMENTATION_URL",
									Value: cfg.MobileSecurityDocumentationURL,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          cfg.MDCContainerName,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4000,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("256Mi"),
									corev1.ResourceCPU:    resource.MustParse("100m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("128Mi"),
									corev1.ResourceCPU:    resource.MustParse("50m"),
								},
							},
						},
						{
							Name:            cfg.OauthProxyContainerName,
							Image:           constants.OauthProxyImage,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "public",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4180,
								},
							},
							Args: []string{
								"--provider=openshift",
								fmt.Sprintf("--client-id=%s", cr.Spec.OAuthClientId),
								fmt.Sprintf("--client-secret=%s", cr.Spec.OAuthClientSecret),
								"--upstream=http://localhost:4000",
								"--http-address=0.0.0.0:4180",
								"--https-address=",
								fmt.Sprintf("--cookie-secret=%s", cookieSecret),
								"--cookie-httponly=false", // we kill the possibility to run MDC on a http route
								"--pass-access-token=true",
								"--scope=user:full",
								"--bypass-auth-for=/about",
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("20m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("32Mi"),
									corev1.ResourceCPU:    resource.MustParse("10m"),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func newMobileClientAdminRoleBinding(cr *mdcv1alpha1.MobileDeveloperConsole) (*rbacv1.RoleBinding, error) {
	name := cr.Name + "-mobileclient-admin"
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      cr.Name,
				Namespace: cr.Namespace,
			},
		},
	}, nil
}

func newMobileClientAdminRole(cr *mdcv1alpha1.MobileDeveloperConsole) (*rbacv1.Role, error) {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cr.Name + "-mobileclient-admin",
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"secrets", "configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"mdc.aerogear.org"},
				Resources: []string{"mobileclients"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"push.aerogear.org"},
				Resources: []string{"pushapplications", "androidvariants", "iosvariants", "webpushvariants"},
				Verbs:     []string{"get", "list", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"mobile-security-service.aerogear.org"},
				Resources: []string{"mobilesecurityserviceapps"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}, nil
}

func newMDCServiceMonitor(cr *mdcv1alpha1.MobileDeveloperConsole) (*monitoringv1.ServiceMonitor, error) {
	labels := map[string]string{
		"monitoring-key": "middleware",
		"name":           "mobile-developer-console",
	}
	matchLabels := map[string]string{
		"internal": "mdc",
	}
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      "mobile-developer-console",
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					Port: "web",
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		},
	}, nil
}

func newMDCPrometheusRule(cr *mdcv1alpha1.MobileDeveloperConsole) (*monitoringv1.PrometheusRule, error) {
	labels := map[string]string{
		"monitoring-key": "middleware",
	}
	critical := map[string]string{
		"severity": "critical",
	}
	warning := map[string]string{
		"severity": "warning",
	}
	sop_url := fmt.Sprintf("https://github.com/aerogear/mobile-developer-console-operator/blob/%s/SOP/SOP-mdc.adoc", version.Version)
	mdcContainerDownAnnotations := map[string]string{
		"description": "The MDC has been down for more than 5 minutes.",
		"summary":     fmt.Sprintf("The mobile-developer-console is down."),
		"sop_url":     sop_url,
	}
	mdcDownAnnotations := map[string]string{
		"description": "The MDC admin console has been down for more than 5 minutes.",
		"summary":     fmt.Sprintf("The mobile-developer-console admin console endpoint has been unavailable for more that 5 minutes."),
		"sop_url":     sop_url,
	}
	mdcPodCPUHighAnnotations := map[string]string{
		"description": "The MDC pod has been at 90% CPU usage for more than 5 minutes",
		"summary":     fmt.Sprintf("The mobile-developer-console is reporting high cpu usage for more that 5 minutes."),
		"sop_url":     sop_url,
	}
	mdcPodMemHighAnnotations := map[string]string{
		"description": "The MDC pod has been at 90% memory usage for more than 5 minutes",
		"summary":     fmt.Sprintf("The mobile-developer-console is reporting high memory usage for more that 5 minutes."),
		"sop_url":     sop_url,
	}
	objectMetaName := util.ObjectMeta(&cr.ObjectMeta, "mdc").Name
	container := "mdc"
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      "mdc-monitoring",
			Labels:    labels,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: "general.rules",
					Rules: []monitoringv1.Rule{
						{
							Alert: "MobileDeveloperConsoleContainerDown",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("absent(kube_pod_container_status_running{namespace=\"%s\",container=\"%s\"}>=1)", cr.Namespace, container),
							},
							For:         "5m",
							Labels:      critical,
							Annotations: mdcContainerDownAnnotations,
						},
						{
							Alert: "MobileDeveloperConsoleDown",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("absent(kube_endpoint_address_available{endpoint=\"%s\"} >= 1)", objectMetaName),
							},
							For:         "5m",
							Labels:      critical,
							Annotations: mdcDownAnnotations,
						},
						{
							Alert: "MobileDeveloperConsolePodCPUHigh",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("(rate(process_cpu_seconds_total{job='%s'}[1m])) > (((kube_pod_container_resource_limits_cpu_cores{namespace='%s',container='%s'})/100)*90)", objectMetaName, cr.Namespace, container),
							},
							For:         "5m",
							Labels:      warning,
							Annotations: mdcPodCPUHighAnnotations,
						},
						{
							Alert: "MobileDeveloperConsolePodMemoryHigh",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("(process_resident_memory_bytes{job='%s'}) > (((kube_pod_container_resource_limits_memory_bytes{namespace='%s',container='%s'})/100)*90)", objectMetaName, cr.Namespace, container),
							},
							For:         "5m",
							Labels:      warning,
							Annotations: mdcPodMemHighAnnotations,
						},
					},
				},
			},
		},
	}, nil
}

func newMDCGrafanaDashboard(cr *mdcv1alpha1.MobileDeveloperConsole) (*integreatlyv1alpha1.GrafanaDashboard, error) {
	labels := map[string]string{
		"monitoring-key": "middleware",
	}
	objectMetaName := util.ObjectMeta(&cr.ObjectMeta, "mdc").Name
	container := "mdc"
	namespace := cr.Namespace
	return &integreatlyv1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "mdc-application",
			Labels:    labels,
		},
		Spec: integreatlyv1alpha1.GrafanaDashboardSpec{
			Name: "mdcapplication.json",
			Json: `{
				"__requires": [
				  {
					"type": "grafana",
					"id": "grafana",
					"name": "Grafana",
					"version": "4.3.2"
				  },
				  {
					"type": "panel",
					"id": "graph",
					"name": "Graph",
					"version": ""
				  },
				  {
					"type": "datasource",
					"id": "prometheus",
					"name": "Prometheus",
					"version": "1.0.0"
				  },
				  {
					"type": "panel",
					"id": "singlestat",
					"name": "Singlestat",
					"version": ""
				  }
				],
				"annotations": {
				  "list": [
					{
					  "builtIn": 1,
					  "datasource": "-- Grafana --",
					  "enable": true,
					  "hide": true,
					  "iconColor": "rgba(0, 211, 255, 1)",
					  "name": "Annotations & Alerts",
					  "type": "dashboard"
					}
				  ]
				},
				"description": "Application metrics",
				"editable": true,
				"gnetId": null,
				"graphTooltip": 0,
				"links": [],
				"panels": [
				  {
					"collapsed": false,
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 0
					},
					"id": 9,
					"panels": [],
					"repeat": null,
					"title": "Uptime",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": true,
					"dashLength": 10,
					"dashes": false,
					"datasource": "Prometheus",
					"fill": 1,
					"gridPos": {
					  "h": 8,
					  "w": 24,
					  "x": 3,
					  "y": 1
					},
					"id": 14,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"nullPointMode": "null",
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": false,
					"steppedLine": true,
					"targets": [
					  {
						"expr": "kube_endpoint_address_available{namespace='` + namespace + `',endpoint='` + objectMetaName + `'}",
						"format": "time_series",
						"intervalFactor": 1,
						"legendFormat": "MDC Application - Uptime",
						"refId": "A"
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "MDC Application - Uptime",
					"tooltip": {
					  "shared": true,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"collapsed": false,
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 9
					},
					"id": 10,
					"panels": [],
					"repeat": null,
					"title": "Resources",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "Prometheus",
					"fill": 1,
					"gridPos": {
					  "h": 8,
					  "w": 24,
					  "x": 0,
					  "y": 10
					},
					"id": 4,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"nullPointMode": "null",
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": false,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "process_virtual_memory_bytes{namespace='` + namespace + `',job='` + objectMetaName + `'}",
						"format": "time_series",
						"hide": false,
						"intervalFactor": 1,
						"legendFormat": "Virtual Memory",
						"refId": "A"
					  },
					  {
						"expr": "process_resident_memory_bytes{namespace='` + namespace + `',job='` + objectMetaName + `'}",
						"format": "time_series",
						"hide": false,
						"intervalFactor": 2,
						"legendFormat": "Memory Usage",
						"refId": "B",
						"step": 2
					  },
					  {
						"expr": "kube_pod_container_resource_limits_memory_bytes{namespace='` + namespace + `',container=` + container + `}",
						"format": "time_series",
						"hide": false,
						"intervalFactor": 2,
						"legendFormat": "Max Memory Allocation",
						"refId": "C",
						"step": 2
					  },
					  {
						"expr": "((kube_pod_container_resource_limits_memory_bytes{namespace='` + namespace + `',container=` + container + `})/100)*90",
						"format": "time_series",
						"hide": false,
						"intervalFactor": 2,
						"legendFormat": "90% of Max Memory Allocation",
						"refId": "D",
						"step": 2
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Memory Usage",
					"tooltip": {
					  "shared": true,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "bytes",
						"label": null,
						"logBase": 2,
						"max": null,
						"min": 0,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": "Prometheus",
					"fill": 1,
					"gridPos": {
					  "h": 8,
					  "w": 24,
					  "x": 0,
					  "y": 18
					},
					"id": 2,
					"legend": {
					  "avg": false,
					  "current": false,
					  "max": false,
					  "min": false,
					  "show": true,
					  "total": false,
					  "values": false
					},
					"lines": true,
					"linewidth": 1,
					"links": [],
					"nullPointMode": "null",
					"options": {},
					"percentage": false,
					"pointradius": 5,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": false,
					"steppedLine": false,
					"targets": [
					  {
						"expr": "sum(rate(process_cpu_seconds_total{namespace='` + namespace + `',job='` + objectMetaName + `'}[1m]))*1000",
						"format": "time_series",
						"interval": "",
						"intervalFactor": 2,
						"legendFormat": "MDC Service - CPU Usage in Millicores",
						"refId": "A",
						"step": 2
					  },
					  {
						"expr": "(kube_pod_container_resource_limits_cpu_cores{namespace='` + namespace + `',container=` + container + `})*1000",
						"format": "time_series",
						"interval": "",
						"intervalFactor": 2,
						"legendFormat": "Maximum Limit of Millicores",
						"refId": "B",
						"step": 2
					  },
					  {
						"expr": "(((kube_pod_container_resource_limits_cpu_cores{namespace='` + namespace + `',container=` + container + `})*1000)/100)*90",
						"format": "time_series",
						"interval": "",
						"intervalFactor": 2,
						"legendFormat": "90% Limit of Millicores",
						"refId": "C",
						"step": 2
					  }
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "CPU Usage",
					"tooltip": {
					  "shared": true,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "buckets": null,
					  "mode": "time",
					  "name": null,
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "short",
						"label": "Millicores",
						"logBase": 10,
						"max": null,
						"min": null,
						"show": true
					  },
					  {
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
					  }
					],
					"yaxis": {
					  "align": false,
					  "alignLevel": null
					}
				  }
				],
				"refresh": "10s",
				"schemaVersion": 18,
				"style": "dark",
				"tags": [],
				"templating": {
				  "list": []
				},
				"time": {
				  "from": "now/d",
				  "to": "now"
				},
				"timepicker": {
				  "refresh_intervals": [
					"5s",
					"10s",
					"30s",
					"1m",
					"5m",
					"15m",
					"30m",
					"1h",
					"2h",
					"1d"
				  ],
				  "time_options": [
					"5m",
					"15m",
					"1h",
					"6h",
					"12h",
					"24h",
					"2d",
					"7d",
					"30d"
				  ]
				},
				"timezone": "browser",
				"title": "MDC Application",
				"uid": "_fSCcUvZk",
				"version": 3
			  }{
				 "annotations": {
				   "list": [
					 {
					   "builtIn": 1,
					   "datasource": "-- Grafana --",
					   "enable": true,
					   "hide": true,
					   "iconColor": "rgba(0, 211, 255, 1)",
					   "name": "Annotations & Alerts",
					   "type": "dashboard"
					 }
				   ]
				 },
				 "description": "Application metrics",
				 "editable": true,
				 "gnetId": null,
				 "graphTooltip": 0,
				 "id": 11,
				 "links": [],
				 "panels": [
				   {
					 "collapsed": false,
					 "gridPos": {
					   "h": 1,
					   "w": 24,
					   "x": 0,
					   "y": 0
					 },
					 "id": 9,
					 "panels": [],
					 "repeat": null,
					 "title": "Uptime",
					 "type": "row"
				   },
				   {
					 "aliasColors": {},
					 "bars": true,
					 "dashLength": 10,
					 "dashes": false,
					 "datasource": "Prometheus",
					 "fill": 1,
					 "gridPos": {
					   "h": 8,
					   "w": 24,
					   "x": 3,
					   "y": 1
					 },
					 "id": 14,
					 "legend": {
					   "avg": false,
					   "current": false,
					   "max": false,
					   "min": false,
					   "show": true,
					   "total": false,
					   "values": false
					 },
					 "lines": true,
					 "linewidth": 1,
					 "links": [],
					 "nullPointMode": "null",
					 "options": {},
					 "percentage": false,
					 "pointradius": 5,
					 "points": false,
					 "renderer": "flot",
					 "seriesOverrides": [],
					 "spaceLength": 10,
					 "stack": false,
					 "steppedLine": true,
					 "targets": [
					   {
						 "expr": "kube_endpoint_address_available{namespace='` + namespace + `',endpoint='` + objectMetaName + `'}",
						 "format": "time_series",
						 "intervalFactor": 1,
						 "legendFormat": "MDC Application - Uptime",
						 "refId": "A"
					   }
					 ],
					 "thresholds": [],
					 "timeFrom": null,
					 "timeRegions": [],
					 "timeShift": null,
					 "title": "MDC Application - Uptime",
					 "tooltip": {
					   "shared": true,
					   "sort": 0,
					   "value_type": "individual"
					 },
					 "type": "graph",
					 "xaxis": {
					   "buckets": null,
					   "mode": "time",
					   "name": null,
					   "show": true,
					   "values": []
					 },
					 "yaxes": [
					   {
						 "format": "short",
						 "label": null,
						 "logBase": 1,
						 "max": null,
						 "min": null,
						 "show": true
					   },
					   {
						 "format": "short",
						 "label": null,
						 "logBase": 1,
						 "max": null,
						 "min": null,
						 "show": true
					   }
					 ],
					 "yaxis": {
					   "align": false,
					   "alignLevel": null
					 }
				   },
				   {
					 "collapsed": false,
					 "gridPos": {
					   "h": 1,
					   "w": 24,
					   "x": 0,
					   "y": 9
					 },
					 "id": 10,
					 "panels": [],
					 "repeat": null,
					 "title": "Resources",
					 "type": "row"
				   },
				   {
					 "aliasColors": {},
					 "bars": false,
					 "dashLength": 10,
					 "dashes": false,
					 "datasource": "Prometheus",
					 "fill": 1,
					 "gridPos": {
					   "h": 8,
					   "w": 24,
					   "x": 0,
					   "y": 10
					 },
					 "id": 4,
					 "legend": {
					   "avg": false,
					   "current": false,
					   "max": false,
					   "min": false,
					   "show": true,
					   "total": false,
					   "values": false
					 },
					 "lines": true,
					 "linewidth": 1,
					 "links": [],
					 "nullPointMode": "null",
					 "options": {},
					 "percentage": false,
					 "pointradius": 5,
					 "points": false,
					 "renderer": "flot",
					 "seriesOverrides": [],
					 "spaceLength": 10,
					 "stack": false,
					 "steppedLine": false,
					 "targets": [
					   {
						 "expr": "process_virtual_memory_bytes{namespace='` + namespace + `',job='` + objectMetaName + `'}",
						 "format": "time_series",
						 "hide": false,
						 "intervalFactor": 1,
						 "legendFormat": "Virtual Memory",
						 "refId": "A"
					   },
					   {
						 "expr": "process_resident_memory_bytes{namespace='` + namespace + `',job='` + objectMetaName + `'}",
						 "format": "time_series",
						 "hide": false,
						 "intervalFactor": 2,
						 "legendFormat": "Memory Usage",
						 "refId": "B",
						 "step": 2
					   },
					   {
						 "expr": "kube_pod_container_resource_limits_memory_bytes{namespace='` + namespace + `',container=` + container + `}",
						 "format": "time_series",
						 "hide": false,
						 "intervalFactor": 2,
						 "legendFormat": "Max Memory Allocation",
						 "refId": "C",
						 "step": 2
					   },
					   {
						 "expr": "((kube_pod_container_resource_limits_memory_bytes{namespace='` + namespace + `',container=` + container + `})/100)*90",
						 "format": "time_series",
						 "hide": false,
						 "intervalFactor": 2,
						 "legendFormat": "90% of Max Memory Allocation",
						 "refId": "D",
						 "step": 2
					   }
					 ],
					 "thresholds": [],
					 "timeFrom": null,
					 "timeRegions": [],
					 "timeShift": null,
					 "title": "Memory Usage",
					 "tooltip": {
					   "shared": true,
					   "sort": 0,
					   "value_type": "individual"
					 },
					 "type": "graph",
					 "xaxis": {
					   "buckets": null,
					   "mode": "time",
					   "name": null,
					   "show": true,
					   "values": []
					 },
					 "yaxes": [
					   {
						 "format": "bytes",
						 "label": null,
						 "logBase": 2,
						 "max": null,
						 "min": 0,
						 "show": true
					   },
					   {
						 "format": "short",
						 "label": null,
						 "logBase": 1,
						 "max": null,
						 "min": null,
						 "show": true
					   }
					 ],
					 "yaxis": {
					   "align": false,
					   "alignLevel": null
					 }
				   },
				   {
					 "aliasColors": {},
					 "bars": false,
					 "dashLength": 10,
					 "dashes": false,
					 "datasource": "Prometheus",
					 "fill": 1,
					 "gridPos": {
					   "h": 8,
					   "w": 24,
					   "x": 0,
					   "y": 18
					 },
					 "id": 2,
					 "legend": {
					   "avg": false,
					   "current": false,
					   "max": false,
					   "min": false,
					   "show": true,
					   "total": false,
					   "values": false
					 },
					 "lines": true,
					 "linewidth": 1,
					 "links": [],
					 "nullPointMode": "null",
					 "options": {},
					 "percentage": false,
					 "pointradius": 5,
					 "points": false,
					 "renderer": "flot",
					 "seriesOverrides": [],
					 "spaceLength": 10,
					 "stack": false,
					 "steppedLine": false,
					 "targets": [
					   {
						 "expr": "sum(rate(process_cpu_seconds_total{namespace='` + namespace + `',job='` + objectMetaName + `'}[1m]))*1000",
						 "format": "time_series",
						 "interval": "",
						 "intervalFactor": 2,
						 "legendFormat": "MDC Service - CPU Usage in Millicores",
						 "refId": "A",
						 "step": 2
					   },
					   {
						 "expr": "(kube_pod_container_resource_limits_cpu_cores{namespace='` + namespace + `',container=` + container + `})*1000",
						 "format": "time_series",
						 "interval": "",
						 "intervalFactor": 2,
						 "legendFormat": "Maximum Limit of Millicores",
						 "refId": "B",
						 "step": 2
					   },
					   {
						 "expr": "(((kube_pod_container_resource_limits_cpu_cores{namespace='` + namespace + `',container=` + container + `})*1000)/100)*90",
						 "format": "time_series",
						 "interval": "",
						 "intervalFactor": 2,
						 "legendFormat": "90% Limit of Millicores",
						 "refId": "C",
						 "step": 2
					   }
					 ],
					 "thresholds": [],
					 "timeFrom": null,
					 "timeRegions": [],
					 "timeShift": null,
					 "title": "CPU Usage",
					 "tooltip": {
					   "shared": true,
					   "sort": 0,
					   "value_type": "individual"
					 },
					 "type": "graph",
					 "xaxis": {
					   "buckets": null,
					   "mode": "time",
					   "name": null,
					   "show": true,
					   "values": []
					 },
					 "yaxes": [
					   {
						 "format": "short",
						 "label": "Millicores",
						 "logBase": 10,
						 "max": null,
						 "min": null,
						 "show": true
					   },
					   {
						 "format": "short",
						 "label": null,
						 "logBase": 1,
						 "max": null,
						 "min": null,
						 "show": true
					   }
					 ],
					 "yaxis": {
					   "align": false,
					   "alignLevel": null
					 }
				   }
				 ],
				 "refresh": "10s",
				 "schemaVersion": 18,
				 "style": "dark",
				 "tags": [],
				 "templating": {
				   "list": []
				 },
				 "time": {
				   "from": "now/d",
				   "to": "now"
				 },
				 "timepicker": {
				   "refresh_intervals": [
					 "5s",
					 "10s",
					 "30s",
					 "1m",
					 "5m",
					 "15m",
					 "30m",
					 "1h",
					 "2h",
					 "1d"
				   ],
				   "time_options": [
					 "5m",
					 "15m",
					 "1h",
					 "6h",
					 "12h",
					 "24h",
					 "2d",
					 "7d",
					 "30d"
				   ]
				 },
				 "timezone": "browser",
				 "title": "MDC Application",
				 "version": 1
			   }`,
		},
	}, nil
}

// reconcileMobileDeveloperRole takes the current state of the Role and ensures it is at the desired
// state. If it already was at the desired state, the OperationResult of the
// controllerutil.CreateOrUpdate call that invokes this will be "unchanged"
func reconcileMobileDeveloperRole(role *rbacv1.Role) {
	role.Rules = []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"mdc.aerogear.org"},
			Resources: []string{"mobileclients"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"push.aerogear.org"},
			Resources: []string{"pushapplications", "androidvariants", "iosvariants", "webpushvariants"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"mobile-security-service.aerogear.org"},
			Resources: []string{"mobilesecurityserviceapps"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"aerogear.org"},
			Resources: []string{"keycloakrealms"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		rbacv1.PolicyRule{
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes"},
			Verbs:     []string{"get", "list"},
		},
	}
}

// reconcileMobileDeveloperRoleBinding takes the current state of the RoleBinding and ensures it is
// at the desired state. If it already was at the desired state, the OperationResult of the
// controllerutil.CreateOrUpdate call that invokes this will be "unchanged"
func reconcileMobileDeveloperRoleBinding(roleBinding *rbacv1.RoleBinding) {
	roleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     "mobile-developer",
	}
	roleBinding.Subjects = []rbacv1.Subject{{
		APIGroup: rbacv1.GroupName,
		Kind:     rbacv1.GroupKind,
		Name:     "system:authenticated",
	}}
}
