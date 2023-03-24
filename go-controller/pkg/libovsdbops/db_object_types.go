package libovsdbops

const (
	addressSet dbObjType = iota
	acl
)

const (
	// owner types
	EgressFirewallDNSOwnerType ownerType = "EgressFirewallDNS"
	EgressQoSOwnerType         ownerType = "EgressQoS"
	// only used for cleanup now, as the stale owner of network policy address sets
	NetworkPolicyOwnerType      ownerType = "NetworkPolicy"
	NetpolDefaultOwnerType      ownerType = "NetpolDefault"
	PodSelectorOwnerType        ownerType = "PodSelector"
	NamespaceOwnerType          ownerType = "Namespace"
	HybridNodeRouteOwnerType    ownerType = "HybridNodeRoute"
	EgressIPOwnerType           ownerType = "EgressIP"
	EgressServiceOwnerType      ownerType = "EgressService"
	MulticastNamespaceOwnerType ownerType = "MulticastNS"
	MulticastClusterOwnerType   ownerType = "MulticastCluster"

	// owner extra IDs, make sure to define only 1 ExternalIDKey for every string value
	PriorityKey           ExternalIDKey = "priority"
	PolicyDirectionKey    ExternalIDKey = "direction"
	GressIdxKey           ExternalIDKey = "gress-index"
	AddressSetIPFamilyKey ExternalIDKey = "ip-family"
	TypeKey               ExternalIDKey = "type"
)

// ObjectIDsTypes should only be created here

var AddressSetEgressFirewallDNS = newObjectIDsType(addressSet, EgressFirewallDNSOwnerType, []ExternalIDKey{
	// dnsName
	ObjectNameKey,
	AddressSetIPFamilyKey,
})

var AddressSetHybridNodeRoute = newObjectIDsType(addressSet, HybridNodeRouteOwnerType, []ExternalIDKey{
	// nodeName
	ObjectNameKey,
	AddressSetIPFamilyKey,
})

var AddressSetEgressQoS = newObjectIDsType(addressSet, EgressQoSOwnerType, []ExternalIDKey{
	// namespace
	ObjectNameKey,
	// egress qos priority
	PriorityKey,
	AddressSetIPFamilyKey,
})

var AddressSetPodSelector = newObjectIDsType(addressSet, PodSelectorOwnerType, []ExternalIDKey{
	// pod selector string representation
	ObjectNameKey,
	AddressSetIPFamilyKey,
})

// deprecated, should only be used for sync
var AddressSetNetworkPolicy = newObjectIDsType(addressSet, NetworkPolicyOwnerType, []ExternalIDKey{
	// namespace_name
	ObjectNameKey,
	// egress or ingress
	PolicyDirectionKey,
	// gress rule index
	GressIdxKey,
	AddressSetIPFamilyKey,
})

var AddressSetNamespace = newObjectIDsType(addressSet, NamespaceOwnerType, []ExternalIDKey{
	// namespace
	ObjectNameKey,
	AddressSetIPFamilyKey,
})

var AddressSetEgressIP = newObjectIDsType(addressSet, EgressIPOwnerType, []ExternalIDKey{
	// cluster-wide address set name
	ObjectNameKey,
	AddressSetIPFamilyKey,
})

var AddressSetEgressService = newObjectIDsType(addressSet, EgressServiceOwnerType, []ExternalIDKey{
	// cluster-wide address set name
	ObjectNameKey,
	AddressSetIPFamilyKey,
})

var ACLNetpolDefault = newObjectIDsType(acl, NetpolDefaultOwnerType, []ExternalIDKey{
	// for now there is only 1 acl of this type, but we use a name in case more types are needed in the future
	ObjectNameKey,
	// egress or ingress
	PolicyDirectionKey,
})

var ACLMulticastNamespace = newObjectIDsType(acl, MulticastNamespaceOwnerType, []ExternalIDKey{
	// namespace
	ObjectNameKey,
	// egress or ingress
	PolicyDirectionKey,
})

var ACLMulticastCluster = newObjectIDsType(acl, MulticastClusterOwnerType, []ExternalIDKey{
	// cluster-scoped multicast acls
	// there are 2 possible TypeKey values for cluster default multicast acl: DefaultDeny and AllowInterNode
	TypeKey,
	// egress or ingress
	PolicyDirectionKey,
})
