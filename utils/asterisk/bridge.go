package asterisk

import (
	"context"

	"github.com/CyCoreSystems/ari"
	"github.com/CyCoreSystems/ari/rid"
)

func CreateBridge(
	ctx context.Context,
	client ari.Client,
	src *ari.Key,
) (*ari.BridgeHandle, error) {

	key := src.New(ari.BridgeKey, rid.New(rid.Bridge))
	bridge, err := client.Bridge().Create(key, "mixing", key.ID)
	if err != nil {
		return nil, err
	}
	return bridge, nil
}
