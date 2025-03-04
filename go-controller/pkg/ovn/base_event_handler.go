package ovn

import (
	"fmt"
	"reflect"

	kapi "k8s.io/api/core/v1"
	knet "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	mnpapi "github.com/k8snetworkplumbingwg/multi-networkpolicy/pkg/apis/k8s.cni.cncf.io/v1beta1"
	egressfirewall "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressfirewall/v1"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/factory"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/util"
)

type baseNetworkControllerEventHandler struct{}

// hasResourceAnUpdateFunc returns true if the given resource type has a dedicated update function.
// It returns false if, upon an update event on this resource type, we instead need to first delete the old
// object and then add the new one.
func hasResourceAnUpdateFunc(objType reflect.Type) bool {
	switch objType {
	case factory.PodType,
		factory.NodeType,
		factory.EgressIPType,
		factory.EgressIPNamespaceType,
		factory.EgressIPPodType,
		factory.EgressNodeType,
		factory.EgressFwNodeType,
		factory.CloudPrivateIPConfigType,
		factory.LocalPodSelectorType,
		factory.NamespaceType,
		factory.MultiNetworkPolicyType:
		return true
	}
	return false
}

// AreResourcesEqual returns true if, given two objects of a known resource type, the update logic for this resource
// type considers them equal and therefore no update is needed. It returns false when the two objects are not considered
// equal and an update needs be executed. This is regardless of how the update is carried out (whether with a dedicated update
// function or with a delete on the old obj followed by an add on the new obj).
func (h *baseNetworkControllerEventHandler) areResourcesEqual(objType reflect.Type, obj1, obj2 interface{}) (bool, error) {
	// switch based on type
	switch objType {
	case factory.PolicyType:
		np1, ok := obj1.(*knet.NetworkPolicy)
		if !ok {
			return false, fmt.Errorf("could not cast obj1 of type %T to *knet.NetworkPolicy", obj1)
		}
		np2, ok := obj2.(*knet.NetworkPolicy)
		if !ok {
			return false, fmt.Errorf("could not cast obj2 of type %T to *knet.NetworkPolicy", obj2)
		}
		return reflect.DeepEqual(np1, np2), nil

	case factory.NodeType:
		node1, ok := obj1.(*kapi.Node)
		if !ok {
			return false, fmt.Errorf("could not cast obj1 of type %T to *kapi.Node", obj1)
		}
		node2, ok := obj2.(*kapi.Node)
		if !ok {
			return false, fmt.Errorf("could not cast obj2 of type %T to *kapi.Node", obj2)
		}

		// when shouldUpdateNode is false, the hostsubnet is not assigned by ovn-kubernetes
		shouldUpdate, err := shouldUpdateNode(node2, node1)
		if err != nil {
			klog.Errorf(err.Error())
		}
		return !shouldUpdate, nil

	case factory.PodType,
		factory.EgressIPPodType:
		// For these types, there was no old vs new obj comparison in the original update code,
		// so pretend they're always different so that the update code gets executed
		return false, nil

	case factory.EgressFirewallType:
		oldEgressFirewall, ok := obj1.(*egressfirewall.EgressFirewall)
		if !ok {
			return false, fmt.Errorf("could not cast obj1 of type %T to *egressfirewall.EgressFirewall", obj1)
		}
		newEgressFirewall, ok := obj2.(*egressfirewall.EgressFirewall)
		if !ok {
			return false, fmt.Errorf("could not cast obj2 of type %T to *egressfirewall.EgressFirewall", obj2)
		}
		return reflect.DeepEqual(oldEgressFirewall.Spec, newEgressFirewall.Spec), nil

	case factory.EgressIPType,
		factory.EgressIPNamespaceType,
		factory.EgressNodeType,
		factory.CloudPrivateIPConfigType:
		// force update path for EgressIP resource.
		return false, nil

	case factory.EgressFwNodeType:
		oldNode, ok := obj1.(*kapi.Node)
		if !ok {
			return false, fmt.Errorf("could not cast obj1 of type %T to *kapi.Node", obj1)
		}
		newNode, ok := obj2.(*kapi.Node)
		if !ok {
			return false, fmt.Errorf("could not cast obj2 of type %T to *kapi.Node", obj2)
		}
		if reflect.DeepEqual(oldNode.Labels, newNode.Labels) &&
			reflect.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses) {
			return true, nil
		}
		return false, nil

	case factory.NamespaceType:
		// force update path for Namespace resource.
		return false, nil

	case factory.MultiNetworkPolicyType:
		mnp1, ok := obj1.(*mnpapi.MultiNetworkPolicy)
		if !ok {
			return false, fmt.Errorf("could not cast obj1 of type %T to *multinetworkpolicyapi.MultiNetworkPolicy", obj1)
		}
		mnp2, ok := obj2.(*mnpapi.MultiNetworkPolicy)
		if !ok {
			return false, fmt.Errorf("could not cast obj2 of type %T to *multinetworkpolicyapi.MultiNetworkPolicy", obj2)
		}
		return reflect.DeepEqual(mnp1, mnp2), nil
	}

	return false, fmt.Errorf("no object comparison for type %s", objType)
}

// Given an object key and its type, getResourceFromInformerCache returns the latest state of the object
// from the informers cache.
func (h *baseNetworkControllerEventHandler) getResourceFromInformerCache(objType reflect.Type, watchFactory *factory.WatchFactory,
	key string) (interface{}, error) {
	var obj interface{}
	var namespace, name string
	var err error

	namespace, name, err = cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to split key %s: %v", key, err)
	}

	switch objType {
	case factory.PolicyType:
		obj, err = watchFactory.GetNetworkPolicy(namespace, name)

	case factory.NodeType,
		factory.EgressNodeType,
		factory.EgressFwNodeType:
		obj, err = watchFactory.GetNode(name)

	case factory.PodType,
		factory.EgressIPPodType:
		obj, err = watchFactory.GetPod(namespace, name)

	case factory.EgressIPNamespaceType,
		factory.NamespaceType:
		obj, err = watchFactory.GetNamespace(name)

	case factory.EgressFirewallType:
		obj, err = watchFactory.GetEgressFirewall(namespace, name)

	case factory.EgressIPType:
		obj, err = watchFactory.GetEgressIP(name)

	case factory.CloudPrivateIPConfigType:
		obj, err = watchFactory.GetCloudPrivateIPConfig(name)

	case factory.MultiNetworkPolicyType:
		obj, err = watchFactory.GetMultiNetworkPolicy(namespace, name)

	default:
		err = fmt.Errorf("object type %s not supported, cannot retrieve it from informers cache",
			objType)
	}
	return obj, err
}

// Given an object and its type, isResourceScheduled returns true if the object has been scheduled.
// Only applied to pods for now. Returns true for all other types.
func (h *baseNetworkControllerEventHandler) isResourceScheduled(objType reflect.Type, obj interface{}) bool {
	switch objType {
	case factory.PodType:
		pod := obj.(*kapi.Pod)
		return util.PodScheduled(pod)
	}
	return true
}

// Given an object type, resourceNeedsUpdate returns true if the object needs to invoke update during iterate retry.
func needsUpdateDuringRetry(objType reflect.Type) bool {
	switch objType {
	case factory.EgressNodeType,
		factory.EgressFwNodeType,
		factory.EgressIPType,
		factory.EgressIPPodType,
		factory.EgressIPNamespaceType,
		factory.CloudPrivateIPConfigType,
		factory.MultiNetworkPolicyType:
		return true
	}
	return false
}

// IsObjectInTerminalState returns true if the object is in a terminal state.
func (h *baseNetworkControllerEventHandler) isObjectInTerminalState(objType reflect.Type, obj interface{}) bool {
	switch objType {
	case factory.PodType,
		factory.EgressIPPodType:
		pod := obj.(*kapi.Pod)
		return util.PodCompleted(pod)

	default:
		return false
	}
}
