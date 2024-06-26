package openrtb_ext

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"

	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

func cloneSupplyChain(schain *openrtb2.SupplyChain) *openrtb2.SupplyChain {
	if schain == nil {
		return nil
	}
	clone := *schain
	clone.Nodes = make([]openrtb2.SupplyChainNode, len(schain.Nodes))
	for i, node := range schain.Nodes {
		clone.Nodes[i] = node
		clone.Nodes[i].HP = ptrutil.Clone(schain.Nodes[i].HP)
	}

	return &clone

}

// SerializeSupplyChain convert schain object to serialized string
func SerializeSupplyChain(schain *openrtb2.SupplyChain) string {

	if len(schain.Nodes) < 1 {
		return ""
	}
	var serializedSchain strings.Builder
	serializedSchain.Grow(256)

	serializedSchain.WriteString(schain.Ver)
	serializedSchain.WriteByte(',')
	fmt.Fprintf(&serializedSchain, "%d", schain.Complete)

	for _, node := range schain.Nodes {
		serializedSchain.WriteByte('!')

		if node.ASI != "" {
			serializedSchain.WriteString(url.QueryEscape(node.ASI))
		}
		serializedSchain.WriteByte(',')

		if node.SID != "" {
			serializedSchain.WriteString(url.QueryEscape(node.SID))
		}
		serializedSchain.WriteByte(',')

		if node.HP != nil {
			// node.HP is integer pointer so 1st dereference it then convert it to string and push to serializedSchain
			fmt.Fprintf(&serializedSchain, "%d", *node.HP)
		}
		serializedSchain.WriteByte(',')

		if node.RID != "" {
			serializedSchain.WriteString(url.QueryEscape(node.RID))
		}
		serializedSchain.WriteByte(',')

		if node.Name != "" {
			serializedSchain.WriteString(url.QueryEscape(node.Name))
		}
		serializedSchain.WriteByte(',')

		if node.Domain != "" {
			serializedSchain.WriteString(url.QueryEscape(node.Domain))
		}
		if node.Ext != nil {
			serializedSchain.WriteByte(',')
			serializedSchain.WriteString(url.QueryEscape(string(node.Ext)))
		}
	}
	return serializedSchain.String()
}
