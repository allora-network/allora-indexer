package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var lastProcessedHeight int64 = 58420    //#25600

type Command struct {
	Parts []string
}

type ClientConfig struct {
	Mode     string
	Node     string
	CliApp   string
	Commands map[string]Command
}
var config ClientConfig

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

	log.Info().Strs("command", completeParts).Msg("Executing command")
	cmd := exec.Command(completeParts[0], completeParts[1:]...)
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

	log.Info().Str("commandName", key).Msg("Starting execution")
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
	// fmt.Printf("%+v\n", result)
	return result, nil
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var (
		modeFlag	   string
		nodeFlag       string
		cliAppFlag     string
		connectionFlag string
	)

	flag.StringVar(&modeFlag, "mode", "blocks", "What to index (blocks, txs or consensus)")
	flag.StringVar(&nodeFlag, "node", "https://allora-rpc.devnet.behindthecurtain.xyz:443", "Node address") //# https://default-node-address:443",
	flag.StringVar(&cliAppFlag, "cliApp", "allorad", "CLI app to execute commands")
	flag.StringVar(&connectionFlag, "conn", "postgres://pump:pump@localhost:5432/pump", "Database connection string")
	flag.Parse()

	// define the commands to execute payloads
	config = ClientConfig{
		Mode:   modeFlag,
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
			// "activeTopicsByHeight": {
			// 	Parts: []string{"{cliApp}", "query", "emissions", "active-topics", "--node", "{node}", "--output", "json", "{\"limit\":\"5\"}", "--height"},
			// },
			"topicById": {
				Parts: []string{"{cliApp}", "query", "emissions", "topic", "--node", "{node}", "--output", "json"},   // Requires "{topic}"
			},
		},
	}

	// Init DB
	initDB(connectionFlag)
	defer closeDB()


	// Initialize the lastProcessedHeight with the latest block height from the database
	// var err error
	// lastProcessedHeight, err = getLatestBlockHeightFromDB()
	// if err != nil {
	// 	log.Error().Err(err).Msg("Failed to get the latest block height from the database")
	// 	return
	// }

	// Fetch and process the latest block once before starting the loop
	// processLatestBlock(config)
	runParsing(config)
	// Set up a ticker to check for new blocks every 4 seconds
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	// Set up a channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			runParsing(config)


		case <-signalChan:
			log.Info().Msg("Shutdown signal received, exiting...")
			// wg.Wait() // Wait for all workers to finish
			return
		}
	}

}

func runParsing(config ClientConfig) {

	switch config.Mode {
	case "txs":
		// Process MsgProcessInferences
		log.Info().Msg("Processing Load topics...")
		// processTopics(config)

		log.Info().Msg("Processing encoded TXs...")
		processTxs(config)
		// Add your processing logic here
	case "consensus":
		// Process MsgProcessInferences
		log.Info().Msg("Processing consensus...")
		// get the latest consensus params in the block
		processConsensusParams(config)
	case "blocks":
		// Process MsgProcessInferences
		log.Info().Msg("Processing blocks...")
		println("Processing blocks...", lastProcessedHeight)
		processLatestBlock(config)
	default:
		log.Info().Msg("Unknown mode, supported modes are blocks, txs, consensus")
	}

}

