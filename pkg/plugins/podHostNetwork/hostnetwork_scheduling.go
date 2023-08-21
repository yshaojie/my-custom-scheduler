package podHostNetwork

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type HostNetwork struct {
	replicaSetLister appsv1.ReplicaSetLister
	deploymentLister appsv1.DeploymentLister
	podLister        listerscorev1.PodLister
}

const (
	Name      = "HostNetwork"
	ErrReason = "HostNetwork port conflict"
)

var _ framework.FilterPlugin = &HostNetwork{}

func (h HostNetwork) Name() string {
	return Name
}

func (h HostNetwork) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if nodeInfo == nil || pod == nil || !pod.Spec.HostNetwork {
		return nil
	}

	//
	//if len(pod.OwnerReferences) == 0 {
	//	return nil
	//}
	//podOwnerReference := pod.OwnerReferences[0]
	//if podOwnerReference.Kind != "ReplicaSet" {
	//	return nil
	//}

	replicaSets, err := h.replicaSetLister.GetPodReplicaSets(pod)
	if err != nil {
		return framework.AsStatus(err)
	}
	if len(replicaSets) == 0 {
		return nil
	}
	deployments, err := h.deploymentLister.Deployments(pod.Namespace).List(labels.Everything())
	if err != nil {
		return framework.AsStatus(err)
	}

	for _, replicaSet := range replicaSets {
		if len(replicaSet.OwnerReferences) == 0 {
			pods, err := h.podLister.Pods(pod.Namespace).List(labels.SelectorFromSet(replicaSet.Labels))
			if err != nil {
				return framework.AsStatus(err)
			}
			if hasHostNetworkPod(pods, nodeInfo) {
				return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReason)
			}
		}
		for _, deployment := range deployments {
			for _, reference := range replicaSet.OwnerReferences {
				if reference.UID == deployment.UID {
					pods, err := h.podLister.Pods(pod.Namespace).List(labels.SelectorFromSet(deployment.Labels))
					if err != nil {
						return framework.AsStatus(err)
					}
					if hasHostNetworkPod(pods, nodeInfo) {
						return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReason)
					}
				}
			}
		}
	}

	return nil

}

func hasHostNetworkPod(ret []*v1.Pod, nodeInfo *framework.NodeInfo) bool {
	if len(ret) == 0 {
		return false
	}
	for _, pod := range ret {
		if pod.Spec.HostNetwork && nodeInfo.Node().Name == pod.Spec.NodeName {
			return true
		}
	}
	return false
}

func New(rawArgs runtime.Object, fh framework.Handle) (framework.Plugin, error) {
	network := HostNetwork{
		replicaSetLister: fh.SharedInformerFactory().Apps().V1().ReplicaSets().Lister(),
		deploymentLister: fh.SharedInformerFactory().Apps().V1().Deployments().Lister(),
		podLister:        fh.SharedInformerFactory().Core().V1().Pods().Lister(),
	}

	return network, nil
}
