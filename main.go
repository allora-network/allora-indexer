package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Command struct {
	Parts []string
}

type ClientConfig struct {
	Node     string
	CliApp   string
	Commands map[string]Command
}

var config ClientConfig
var workersNum uint

func ExecuteCommand(cliApp, node string, parts []string) ([]byte, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("no command parts provided")
	}

	var completeParts []string
	for _, part := range parts {
		completeParts = append(completeParts, part)
	}

	completeParts = replacePlaceholders(completeParts, "{node}", node)
	completeParts = replacePlaceholders(completeParts, "{cliApp}", cliApp)

	log.Debug().Strs("command", completeParts).Msg("Executing command")
	cmd := exec.Command(completeParts[0], completeParts[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Command execution failed")
		return nil, err
	}

	return output, nil
}

func replacePlaceholders(parts []string, placeholder, value string) []string {
	var replacedParts []string
	for _, part := range parts {
		if part == placeholder {
			replacedParts = append(replacedParts, value)
		} else {
			replacedParts = append(replacedParts, part)
		}
	}
	return replacedParts
}

func ExecuteCommandByKey[T any](config ClientConfig, key string, params ...string) (T, error) {
	var result T

	cmd, ok := config.Commands[key]
	if !ok {
		return result, fmt.Errorf("command not found")
	}

	if len(params) > 0 {
		cmd.Parts = append(cmd.Parts, params...)
	}

	log.Debug().Str("commandName", key).Msg("Starting execution")
	output, err := ExecuteCommand(config.CliApp, config.Node, cmd.Parts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		return result, err
	}

	log.Debug().Str("raw output", string(output)).Msg("Command raw output")

	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Error().Err(err).Str("json", string(output)).Msg("Failed to unmarshal JSON")
		return result, err
	}

	return result, nil
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var (
		nodeFlag         string
		cliAppFlag       string
		connectionFlag   string
		exitWhenCaughtUp bool
		blocks           []string
	)

	pflag.UintVar(&workersNum, "workersNum", 5, "Number of workers to process blocks concurrently")
	pflag.StringVar(&nodeFlag, "node", "https://allora-rpc.edgenet.allora.network:443/", "Node address") //# https://default-node-address:443",
	pflag.StringVar(&cliAppFlag, "cliApp", "allorad", "CLI app to execute commands")
	pflag.StringVar(&connectionFlag, "conn", "postgres://pump:pump@localhost:5432/pump", "Database connection string")
	pflag.StringSliceVar(&blocks, "blocks", nil, "A list of blocks to process.")
	pflag.BoolVar(&exitWhenCaughtUp, "exitWhenCaughtUp", true, "Exit when last block is processed. If false will keep processing new blocks.")
	pflag.Parse()

	// define the commands to execute payloads
	config = ClientConfig{
		Node:   nodeFlag,
		CliApp: cliAppFlag,
		Commands: map[string]Command{
			"latestBlock": {
				Parts: []string{"{cliApp}", "query", "consensus", "comet", "block-latest", "--node", "{node}", "--output", "json"},
			},
			"blockByHeight": { // Add a template command for fetching blocks by height
				Parts: []string{"{cliApp}", "query", "block", "--type=height", "--node", "{node}", "--output", "json", "{height}"},
			},
			"consensusParams": {
				Parts: []string{"{cliApp}", "query", "consensus", "params", "--node", "{node}", "--output", "json"},
			},
			"decodeTx": {
				Parts: []string{"{cliApp}", "tx", "decode", "--node", "{node}", "--output", "json"}, // Requires , "{txData}"
			},
			"nextTopicId": {
				Parts: []string{"{cliApp}", "query", "emissions", "next-topic-id", "--node", "{node}", "--output", "json"},
			},
			"topicById": {
				Parts: []string{"{cliApp}", "query", "emissions", "topic", "--node", "{node}", "--output", "json"}, // Requires "{topic}"
			},
		},
	}

	// Init DB
	initDB(connectionFlag)
	defer closeDB()

	// Set up a channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Set a cancel context to stop the workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a channel to listen for block heights to process
	heightsChan := make(chan uint64, workersNum)
	defer close(heightsChan)

	wgBlocks := sync.WaitGroup{}
	for j := uint(1); j <= workersNum; j++ {
		wgBlocks.Add(1)
		go worker(ctx, &wgBlocks, heightsChan)
	}
	defer wgBlocks.Wait() // Wait for all workers to finish at the end of the main function

	if len(blocks) > 0 {
		log.Info().Msgf("Processing only particular blocks: %v", blocks)
		for _, block := range blocks {
			height, err := strconv.ParseUint(block, 10, 64)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to parse block height: %s", block)
			}
			heightsChan <- height
		}
	} else {
		// If no blocks are provided, start the main loop
		log.Info().Msg("Starting main loop...")
		generateBlocksLoop(ctx, signalChan, heightsChan, exitWhenCaughtUp)
	}

	log.Info().Msg("Exited main loop, waiting for subroutines to finish...")
	cancel()
}

// Generates the block heights to process in an infinite loop
func generateBlocksLoop(ctx context.Context, signalChan <-chan os.Signal, heightsChan chan<- uint64, exitWhenCaughtUp bool) {
	for {
		lastProcessedHeight, err := getLatestBlockHeightFromDB()
		if err != nil {
			log.Error().Err(err).Msg("Failed to getLatestBlockHeightFromDB")
		}
		chainLatestHeight, err := getLatestHeight()
		if err != nil {
			log.Error().Err(err).Msg("Failed to getLatestHeight from chain")
		}
		log.Info().Msgf("Processing heights from %d to %d", lastProcessedHeight, chainLatestHeight)
		// Emit heights to process into channel
		for w := lastProcessedHeight; w <= chainLatestHeight; w++ {
			select {
			case <-signalChan:
				log.Info().Msg("Shutdown signal received, exiting...")
				return
			case <-ctx.Done():
				log.Info().Msg("Context cancelled, exiting...")
				return
			default:
				heightsChan <- w
			}
		}
		log.Info().Msg("All blocks processed...")
		if exitWhenCaughtUp {
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func worker(ctx context.Context, wgBlocks *sync.WaitGroup, heightsChan <-chan uint64) {
	defer wgBlocks.Done()
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, stop the worker
			return
		case height, ok := <-heightsChan:
			if !ok {
				// heightsChan was closed, stop the worker
				log.Warn().Msg("heightsChan closed, stopping worker")
				return
			}
			// Tx
			log.Info().Msgf("Processing height: %d", height)
			block, err := fetchBlock(config, height)
			if err != nil {
				log.Error().Err(err).Msg("Worker: Failed to fetchBlock block height")
			}
			log.Info().Msgf("fetchBlock height: %d, len(TXs): %d", height, len(block.Data.Txs))
			err = writeBlock(config, block)
			if err != nil {
				log.Error().Err(err).Msgf("Worker: Failed to writeBlock, height: %d", height)
			}

			log.Info().Msgf("Write height: %d", height)

			if len(block.Data.Txs) > 0 {
				log.Info().Msgf("Processing txs at height: %d", height)
				wgTxs := sync.WaitGroup{}
				for _, encTx := range block.Data.Txs {
					wgTxs.Add(1)
					log.Info().Msgf("Processing height: %d, encTx: %s", height, encTx)
					go processTx(&wgTxs, height, encTx)
				}
				wgTxs.Wait()
			}

			// Events
			log.Info().Msgf("Processing height: %d", height)
			err = processBlock(config, height)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get block events")
			}
		}
	}

}
