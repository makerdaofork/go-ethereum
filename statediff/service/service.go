package service

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	b "github.com/ethereum/go-ethereum/statediff/builder"
	e "github.com/ethereum/go-ethereum/statediff/extractor"
	p "github.com/ethereum/go-ethereum/statediff/publisher"
	"log"
)

type StateDiffService struct {
	builder    *b.Builder
	extractor  e.Extractor
	blockchain *core.BlockChain
}

func NewStateDiffService(db ethdb.Database, blockChain *core.BlockChain) (*StateDiffService, error) {
	config := statediff.Config{}
	builder := b.NewBuilder(db)
	publisher, err := p.NewPublisher(config)
	if err != nil {
		return nil, nil
	}

	extractor, _ := e.NewExtractor(builder, publisher)
	return &StateDiffService{
		blockchain: blockChain,
		extractor:  extractor,
	}, nil
}

func (StateDiffService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}

}

func (StateDiffService) APIs() []rpc.API {
	return []rpc.API{}
}

func (sds *StateDiffService) loop(sub event.Subscription, events chan core.ChainHeadEvent) {
	defer sub.Unsubscribe()

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				log.Fatalf("Error getting chain head event from subscription.")
			}
			log.Println("doing something with an event", ev)
			previousBlock := ev.Block
			//TODO: figure out the best way to get the previous block
			currentBlock := ev.Block
			sds.extractor.ExtractStateDiff(*previousBlock, *currentBlock)
		}
	}

}
func (sds *StateDiffService) Start(server *p2p.Server) error {
	events := make(chan core.ChainHeadEvent, 10)
	sub := sds.blockchain.SubscribeChainHeadEvent(events)

	go sds.loop(sub, events)

	return nil
}
func (StateDiffService) Stop() error {
	return nil
}