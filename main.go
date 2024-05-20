package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
var workersNum	   uint

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
		nodeFlag       string
		cliAppFlag     string
		connectionFlag string
		exitWhenCaughtUp bool
	)

	flag.UintVar(&workersNum, "workersNum", 5, "Number of workers to process blocks concurrently")
	flag.StringVar(&nodeFlag, "node", "https://allora-rpc.devnet.behindthecurtain.xyz:443", "Node address") //# https://default-node-address:443",
	flag.StringVar(&cliAppFlag, "cliApp", "allorad", "CLI app to execute commands")
	flag.StringVar(&connectionFlag, "conn", "postgres://pump:pump@localhost:5432/pump", "Database connection string")
	flag.BoolVar(&exitWhenCaughtUp, "exitWhenCaughtUp", true, "Exit when last block is processed. If false will keep processing new blocks.")
	flag.Parse()

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
				Parts: []string{"{cliApp}", "tx", "decode", "--node", "{node}", "--output", "json"},   // Requires , "{txData}"
			},
			"nextTopicId": {
				Parts: []string{"{cliApp}", "query", "emissions", "next-topic-id", "--node", "{node}", "--output", "json"},
			},
			"topicById": {
				Parts: []string{"{cliApp}", "query", "emissions", "topic", "--node", "{node}", "--output", "json"},   // Requires "{topic}"
			},
		},
	}

	// Init DB
	initDB(connectionFlag)
	defer closeDB()

	// Set up a channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	wgBlocks := sync.WaitGroup{}

	// Set up a channel to listen for block heights to process
    heightsChan := make(chan uint64, workersNum)

	for j := uint(1); j <= workersNum; j++ {
		wgBlocks.Add(1)
		go worker(&wgBlocks, heightsChan)
	}

	// Emit heights to process into channel
	// Todo we can listen for new blocks via websocket and emit them to the channel
	for{
		lastProcessedHeight, err := getLatestBlockHeightFromDB()
		if err != nil {
			log.Error().Err(err).Msg("Failed to getLatestBlockHeightFromDB")
			return
		}
		chainLatestHeight, err := getLatestHeight()
		if err != nil {
			log.Error().Err(err).Msg("Failed to getLatestHeight")
			return
		}
		log.Info().Msgf("Processing heights from %d to %d", lastProcessedHeight, chainLatestHeight)
		// Emit heights to process into channel
		for w := lastProcessedHeight; w <= chainLatestHeight; w++ {
			select {
			case <-signalChan:
				log.Info().Msg("Shutdown signal received, exiting...")
				// cancel()
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
	close(heightsChan)
	wgBlocks.Wait() // Wait for all workers to finish
	log.Info().Msg("All workers finished")

}

func worker(wgBlocks *sync.WaitGroup, heightsChan <-chan uint64){
	defer wgBlocks.Done()
	for height := range heightsChan {
		log.Info().Msgf("Processing height: %d", height)
		block, err := fetchBlock(config, height)
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetchBlock block height")
			return
		}
		log.Info().Msgf("fetchBlock height: %d, len(TXs): %d", height, len(block.Data.Txs))
		err = writeBlock(config, block)
		if err != nil {
			log.Error().Err(err).Msg("Failed to writeBlock block height")
			return
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
	}
}