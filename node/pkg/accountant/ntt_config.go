package accountant

import (
	"fmt"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole/sdk"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type emitterConfigEntry struct {
	chainId vaa.ChainID
	addr    string
	logOnly bool
}

type emitterConfig []emitterConfigEntry

// nttGetEmitters returns the set of direct NTT and AR emitters based on the environment passed in.
func nttGetEmitters(env common.Environment) (validEmitters, validEmitters, error) {
	var directEmitterConfig emitterConfig
	arEmitterConfig := sdk.KnownAutomaticRelayerEmitters
	if env == common.MainNet {
		directEmitterConfig = emitterConfig{}
	} else if env == common.TestNet {
		directEmitterConfig = emitterConfig{
			{chainId: vaa.ChainIDSolana, addr: "7e6436b671cce379a1fa9833783e28b36d39a00e2cdc6bfeab5d2d836eb61c7f"},
			{chainId: vaa.ChainIDSepolia, addr: "0000000000000000000000001fdc902e30b188fd2ba976b421cb179943f57896"},
			{chainId: vaa.ChainIDArbitrumSepolia, addr: "0000000000000000000000000e24d17d7467467b39bf64a9dff88776bd6c74d7"},
			{chainId: vaa.ChainIDBaseSepolia, addr: "0000000000000000000000001e072169541f1171e427aa44b5fd8924bee71b0e"},
			{chainId: vaa.ChainIDOptimismSepolia, addr: "00000000000000000000000041265eb2863bf0238081f6aeefef73549c82c3dd"},
		}
		arEmitterConfig = sdk.KnownTestnetAutomaticRelayerEmitters
	} else {
		// Every other environment uses the devnet ones.
		directEmitterConfig = emitterConfig{
			{chainId: vaa.ChainIDEthereum, addr: "000000000000000000000000855FA758c77D68a04990E992aA4dcdeF899F654A"},
			{chainId: vaa.ChainIDEthereum, addr: "000000000000000000000000fA2435Eacf10Ca62ae6787ba2fB044f8733Ee843"},
			{chainId: vaa.ChainIDBSC, addr: "000000000000000000000000fA2435Eacf10Ca62ae6787ba2fB044f8733Ee843"},
			{chainId: vaa.ChainIDBSC, addr: "000000000000000000000000855FA758c77D68a04990E992aA4dcdeF899F654A"},
		}
		arEmitterConfig = sdk.KnownDevnetAutomaticRelayerEmitters
	}

	// Build the direct emitter map, setting the payload based on whether or not the config says it should be log only.
	directEmitters := make(validEmitters)
	for _, emitter := range directEmitterConfig {
		addr, err := vaa.StringToAddress(emitter.addr)
		if err != nil {
			return nil, nil, fmt.Errorf(`failed to parse direct emitter address "%s": %w`, emitter.addr, err)
		}
		ek := emitterKey{emitterChainId: emitter.chainId, emitterAddr: addr}
		if _, exists := directEmitters[ek]; exists {
			return nil, nil, fmt.Errorf(`duplicate direct emitter "%s:%s"`, emitter.chainId.String(), emitter.addr)
		}
		directEmitters[ek] = !emitter.logOnly
	}

	// Build the automatic relayer emitter map based on the standard config in the SDK.
	arEmitters := make(validEmitters)
	for _, emitter := range arEmitterConfig {
		addr, err := vaa.StringToAddress(emitter.Addr)
		if err != nil {
			return nil, nil, fmt.Errorf(`failed to parse AR emitter address "%s": %w`, emitter.Addr, err)
		}
		ek := emitterKey{emitterChainId: emitter.ChainId, emitterAddr: addr}
		if _, exists := directEmitters[ek]; exists {
			return nil, nil, fmt.Errorf(`duplicate AR emitter "%s:%s"`, emitter.ChainId.String(), emitter.Addr)
		}
		arEmitters[ek] = true
	}

	return directEmitters, arEmitters, nil
}
