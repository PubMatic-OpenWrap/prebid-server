package openrtb_ext

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

const (
	SChainVersion1             = "1.0"
	SChainNodeFieldsWithoutExt = 6
	SChainNodeFieldsWithExt    = 7
	SChainMetadataCount        = 2
	SChainRequiredLength       = 2
	SChainCompleteYes          = 1
	SChainCompleteNo           = 0
	SIDLength                  = 64
	HPOne                      = 1
)

const (
	ASIIndex = iota
	SIDIndex
	HPIndex
	RIDIndex
	NameIndex
	DomainIndex
	ExtIndex
)

const (
	VersionIndex = iota
	CompleteIndex
)

const (
	MetadataIndex = iota
	NodesStartIndex
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

// DeserializeSupplyChain deserializes the serialized supply chain value into an openrtb2.SupplyChain object.
// It splits the serialized value into individual nodes, parses the remaining fields of sChain from the first node value,
// validates the sChain version, assigns the parsed values to the sChain object,
//
// Algorithm:
//  1. Split the serialized value into individual nodes using the "!" separator.
//  2. Parse the version and complete fields from the first node value.
//  3. Iterate over the remaining node values and split each node value into individual fields using the "," separator
//     and parse the asi, sid, hp, rid, name, and domain fields.
func DeserializeSupplyChain(serializedSChain string) (*openrtb2.SupplyChain, error) {
	if serializedSChain == "" {
		return nil, errors.New("empty schain value")
	}
	// Split the serialized value into individual nodes
	nodeValues := strings.Split(serializedSChain, "!")
	if len(nodeValues) < SChainRequiredLength {
		return nil, fmt.Errorf("invalid schain value | schain value should have schain object and schain nodes")
	}

	// Parse the remaining fields of sChain from the first node value
	sChainObjectValues := strings.Split(nodeValues[MetadataIndex], ",")
	if len(sChainObjectValues) != SChainMetadataCount {
		return nil, fmt.Errorf("invalid schain value | invalid schain object metadata")
	}

	sChain := &openrtb2.SupplyChain{}

	sChain.Ver = sChainObjectValues[VersionIndex]

	sChain.Complete = 0

	if sChainObjectValues[CompleteIndex] != "" {
		complete, err := strconv.Atoi(sChainObjectValues[CompleteIndex])
		if err != nil {
			return nil, fmt.Errorf("unable to convert [%s] to integer", sChainObjectValues[CompleteIndex])
		}
		sChain.Complete = int8(complete)
	}

	sChain.Nodes = make([]openrtb2.SupplyChainNode, 0, len(nodeValues))
	// Parse and add each node to the sChain.Nodes slice
	for _, sChainNode := range nodeValues[NodesStartIndex:] {
		node, err := deserializeSupplyChainNode(sChainNode, serializedSChain)
		if err != nil {
			return nil, err
		}
		sChain.Nodes = append(sChain.Nodes, node)
	}
	return sChain, nil
}

// deserializeSupplyChainNode deserializes a single supply chain node value into an openrtb2.SupplyChainNode object
func deserializeSupplyChainNode(sChainNode, serializedSChain string) (openrtb2.SupplyChainNode, error) {
	fields := strings.Split(sChainNode, ",")
	if len(fields) < SChainNodeFieldsWithoutExt || len(fields) > SChainNodeFieldsWithExt { // fields can have 7 values when ext is present
		return openrtb2.SupplyChainNode{}, fmt.Errorf("invalid schain value | invalid schain node fields")
	}

	asi, err := url.QueryUnescape(fields[ASIIndex])
	if err != nil {
		return openrtb2.SupplyChainNode{}, fmt.Errorf("invalid schain node value: %s | invalid schain node, failed to unescape asi: %v", fields[ASIIndex], err)
	}
	sid, err := url.QueryUnescape(fields[SIDIndex])
	if err != nil {
		return openrtb2.SupplyChainNode{}, fmt.Errorf("invalid schain node value: %s | invalid schain node, failed to unescape sid: %v", fields[SIDIndex], err)
	}
	rid, err := url.QueryUnescape(fields[RIDIndex])
	if err != nil {
		return openrtb2.SupplyChainNode{}, fmt.Errorf("invalid schain node value: %s | invalid schain node, failed to unescape rid: %v", fields[RIDIndex], err)
	}
	name, err := url.QueryUnescape(fields[NameIndex])
	if err != nil {
		return openrtb2.SupplyChainNode{}, fmt.Errorf("invalid schain node value: %s | invalid schain node, failed to unescape name: %v", fields[NameIndex], err)
	}
	domain, err := url.QueryUnescape(fields[DomainIndex])
	if err != nil {
		return openrtb2.SupplyChainNode{}, fmt.Errorf("invalid schain node value: %s | invalid schain node, failed to unescape domain: %v", fields[DomainIndex], err)
	}

	// Convert the hp field to an int64
	hp, err := strconv.Atoi(fields[HPIndex])
	if err != nil {
		return openrtb2.SupplyChainNode{}, fmt.Errorf("unable to convert [%s] to integer", fields[HPIndex])
	}

	var ext json.RawMessage
	if len(fields) == SChainNodeFieldsWithExt {
		ext = json.RawMessage(fields[ExtIndex])
		decodedExt, err := url.QueryUnescape(string(ext))
		if err == nil {
			ext = json.RawMessage(decodedExt)
		}
	}

	// Create and return the supply chain node object
	return openrtb2.SupplyChainNode{
		ASI:    asi,
		SID:    sid,
		HP:     ptrutil.ToPtr(int8(hp)),
		RID:    rid,
		Name:   name,
		Domain: domain,
		Ext:    ext,
	}, nil
}
