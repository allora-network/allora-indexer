package main

import (
	"context"
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

// const workersNum = 10        // Number of jobs to process
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

	log.Info().Strs("command", completeParts).Msg("Executing command")
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
		nodeFlag       string
		cliAppFlag     string
		connectionFlag string
	)

	flag.UintVar(&workersNum, "workersNum", 5, "Number of workers to process blocks concurrently")
	flag.StringVar(&nodeFlag, "node", "https://allora-rpc.devnet.behindthecurtain.xyz:443", "Node address") //# https://default-node-address:443",
	flag.StringVar(&cliAppFlag, "cliApp", "allorad", "CLI app to execute commands")
	flag.StringVar(&connectionFlag, "conn", "postgres://pump:pump@localhost:5432/pump", "Database connection string")
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
	// runParsing(config)
	// Set up a ticker to check for new blocks every 4 seconds
	// ticker := time.NewTicker(4 * time.Second)
	// defer ticker.Stop()

	// Set up a channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

    heightsChan := make(chan uint64, workersNum)
    // results := make(chan int, workersNum)

	// var err error
	// var exit bool = false

	ctx, cancel := context.WithCancel(context.Background())
	wgBlocks := sync.WaitGroup{}

	var height uint64
    
	for j := uint(1); j <= workersNum; j++ {
        // jobs <- j

		go func() {
			for height = range heightsChan {
				wgBlocks.Add(1)

				println("Processing height: ", height)
				block, err := fetchBlock(config, height)
				if err != nil {
					log.Error().Err(err).Msg("Failed to fetchBlock block height")
					break
				}
				err = writeBlock(config, block)
				if err != nil {
					log.Error().Err(err).Msg("Failed to writeBlock block height")
					break
				}

				if len(block.Data.Txs) > 0 {
					println("Processing txs at height: ", height)
					time.Sleep(2 * time.Second)
					wgTxs := sync.WaitGroup{}
					for _, encTx := range block.Data.Txs {
						wgTxs.Add(1)
						println("Processing height, encTx: ", height, encTx)
						go processTx(&wgTxs, height, encTx)
					}
					wgTxs.Wait()

					time.Sleep(10 * time.Second)
				}
				wgBlocks.Done()

				select {
				case <-ctx.Done():
					break
				default:
					continue
				}
			}
		}()
	}

	go func() {
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
		println("lastProcessedHeight: ", lastProcessedHeight,	", chainLatestHeight: ", chainLatestHeight)
	
		for w := lastProcessedHeight; w <= chainLatestHeight; w++ {
			select {
			case <-ctx.Done():
				break
			default:
				println("Sending height to heightsChan", w)
				heightsChan <- w
			}
		}
		time.Sleep(5 * time.Second)

	}()


    // close(jobs)

    // for a := 1; a <= workersNum; a++ {
    //     <-results
    // }



	for {
		select {


		// case height := <-heightsChan:
		// 	wgBlocks.Add(1)
		// 	go func() {
		// 		defer wgBlocks.Done()
		// 		block, err := fetchBlock(config, height)
		// 		if err != nil {
		// 			log.Error().Err(err).Msg("Failed to fetchBlock block height")
		// 			return
		// 		}
		// 		err = writeBlock(config, block)

		// 		wgTxs.Add(len(block.Data.Txs))
		// 		for _, encTx := range block.Data.Txs {
		// 			go processTx(&wgTxs, height, encTx)
		// 		}
		// 		wgTxs.Wait()

		// 	}()


		// case <-ticker.C:
		// 	runParsing(config)


		case <-signalChan:
			log.Info().Msg("Shutdown signal received, exiting...")
			cancel()

			wgBlocks.Wait() // Wait for all workers to finish
			close(heightsChan)
			return
		}
	}

}

// func runParsing(config ClientConfig) {

// 	switch config.Mode {
// 	case "txs":
// 		// Process MsgProcessInferences
// 		log.Info().Msg("Processing Load topics...")
// 		// processTopics(config)

// 		log.Info().Msg("Processing encoded TXs...")
// 		processTxs(config)
// 		// Add your processing logic here
// 	case "consensus":
// 		// Process MsgProcessInferences
// 		log.Info().Msg("Processing consensus...")
// 		// get the latest consensus params in the block
// 		processConsensusParams(config)
// 	case "blocks":
// 		// Process MsgProcessInferences
// 		log.Info().Msg("Processing blocks...")
// 		// println("Processing blocks...", lastProcessedHeight)
// 		processLatestBlock(config)
// 	default:
// 		log.Info().Msg("Unknown mode, supported modes are blocks, txs, consensus")
// 	}

// }

