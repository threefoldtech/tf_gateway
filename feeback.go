package tfgateway

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/client"
	"github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg/provision"
)

type ExplorerFeedback struct {
	client    *client.Client
	converter provision.ResultConverterFunc
}

func NewExplorerFeedback(client *client.Client, converter provision.ResultConverterFunc) *ExplorerFeedback {
	if converter == nil {
		converter = ToSchemaType
	}
	return &ExplorerFeedback{
		client:    client,
		converter: converter,
	}
}

func (e *ExplorerFeedback) Feedback(nodeID string, r *provision.Result) error {
	wr, err := e.converter(*r)
	if err != nil {
		return fmt.Errorf("failed to convert result into schema type: %w")
	}

	return e.client.Workloads.WorkloadPutResult(nodeID, r.ID, *wr)
}

func (e *ExplorerFeedback) Deleted(nodeID, id string) error {
	return e.client.Workloads.WorkloadPutDeleted(nodeID, id)
}

const gib = 1024 * 1024 * 1024

func (e *ExplorerFeedback) UpdateReservedResources(nodeID string, c provision.Counters) error {
	resources := directory.ResourceAmount{
		Cru: c.CRU.Current(),
		Mru: float64(c.MRU.Current()) / float64(gib),
		Sru: float64(c.SRU.Current()) / float64(gib),
		Hru: float64(c.HRU.Current()) / float64(gib),
	}

	workloads := directory.WorkloadAmount{
		// 	Proxy: ,
		// ReverseProxy: ,
		// Subdomain: ,
		// DelegateDomain: ,
		// Volume:       uint16(c.volumes.Current()),
		// Container:    uint16(c.containers.Current()),
		// ZDBNamespace: uint16(c.zdbs.Current()),
		// K8sVM:        uint16(c.vms.Current()),
		// Network:      uint16(c.networks.Current()),
	}
	log.Info().Msgf("reserved resource %+v", resources)
	log.Info().Msgf("provisionned workloads %+v", workloads)
	return e.client.Directory.NodeUpdateUsedResources(nodeID, resources, workloads)
}

// ToSchemaType converts result to schema type
func ToSchemaType(r provision.Result) (*workloads.Result, error) {
	var rType workloads.ResultCategoryEnum
	switch r.Type {
	case ProxyReservation:
		rType = workloads.ResultCategoryProxy
	case ReverseProxyReservation:
		rType = workloads.ResultCategoryReverseProxy
	case SubDomainReservation:
		rType = workloads.ResultCategorySubDomain
	case DomainDeleateReservation:
		rType = workloads.ResultCategoryDomainDelegate
	default:
		return nil, fmt.Errorf("unknown reservation type: %s", r.Type)
	}

	result := workloads.Result{
		Category:   rType,
		WorkloadId: r.ID,
		DataJson:   r.Data,
		Signature:  r.Signature,
		State:      workloads.ResultStateEnum(r.State),
		Message:    r.Error,
		Epoch:      schema.Date{Time: r.Created},
	}

	return &result, nil
}
