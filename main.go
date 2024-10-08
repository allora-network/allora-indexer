package main

import (
	"context"
	"os"
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
var awsAccessKey string
var awsSecretKey string
var s3BucketName string
var s3FileKey string
var parallelJobs uint
var mode string
var bootstrapBlockHeight int64
var maxConcurrentTxPerRoutine uint

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("Application error")
		os.Exit(1)
	}
}

func run() error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var (
		nodeFlag         string
		cliAppFlag       string
		connectionFlag   string
		exitWhenCaughtUp bool
		blocks           []string
	)

	pflag.UintVar(&workersNum, "WORKERS_NUM", 1, "Number of workers to process blocks concurrently")
	pflag.StringVar(&nodeFlag, "NODE", "https://allora-rpc.testnet.allora.network/", "Node address") //# https://default-node-address:443",
	pflag.StringVar(&cliAppFlag, "CLIAPP", "allorad", "CLI app to execute commands")
	pflag.StringVar(&connectionFlag, "CONNECTION", "postgres://app:app@localhost:5433/app", "Database connection string")
	pflag.StringVar(&awsAccessKey, "AWS_ACCESS_KEY", "", "AWS access key")
	pflag.StringVar(&awsSecretKey, "AWS_SECURITY_KEY", "", "AWS security key")
	pflag.StringVar(&s3BucketName, "S3_BUCKET_NAME", "allora-testnet-1-indexer-backups", "AWS s3 bucket name")
	pflag.StringVar(&s3FileKey, "S3_FILE_KEY", "filename.dump", "AWS s3 file key")
	pflag.StringVar(&mode, "MODE", "full", "Mode: 'full' for full update, 'dump' to simply overwrite loading a dump and exit, 'empty' to create empty DB and continue")
	pflag.UintVar(&parallelJobs, "RESTORE_PARALLEL_JOBS", 4, "Number of parallel jobs (workers) to restore the dump")
	pflag.BoolVar(&exitWhenCaughtUp, "EXIT_APP", false, "Exit when last block is processed. If false will keep processing new blocks.")
	pflag.Int64Var(&bootstrapBlockHeight, "BOOTSTRAP_BLOCKHEIGHT", 0, "Start synchronizing on an empty db from this block height - if 0, do not use")
	pflag.UintVar(&maxConcurrentTxPerRoutine, "MAX_CONCURRENT_TX_PROCESSING", 32, "Number of max concurrent routines to process tx")

	pflag.Parse()

	log.Info().
		Uint("WORKERS_NUM", workersNum).
		Str("NODE", nodeFlag).
		Str("CLIAPP", cliAppFlag).
		Str("S3_BUCKET_NAME", s3BucketName).
		Str("S3_FILE_KEY", s3FileKey).
		Str("MODE", mode).
		Bool("EXIT_APP", exitWhenCaughtUp).
		Int64("BOOTSTRAP_BLOCKHEIGHT", bootstrapBlockHeight).
		Uint("MAX_CONCURRENT_TX_PROCESSING", maxConcurrentTxPerRoutine).
		Msg("Allora Indexer started")

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
	err := initDB(connectionFlag)
	if err != nil {
		log.Error().Err(err).Msg("Failed to init DB")
		return err
	}
	defer closeDB()

	if mode == "empty" {
		err := setupDB()
		if err != nil {
			log.Error().Err(err).Msg("Failed to setup DB")
			return err
		}
	} else if mode == "dump" {
		_, err := restoreBackupFromS3()
		if err != nil {
			log.Err(err).Msg("Failed restoring DB and start fetching blockchain data from scratch")
			return err
		}
	} else if mode == "full" {
		blockInfoExists, err := tableExists("block_info")
		if err != nil {
			log.Error().Err(err).Msg("Failed to check if table block_info exists")
			return err
		}
		needsDownloadBackupFromS3 := false
		if !blockInfoExists {
			log.Info().Msg("DB is empty - restoring from S3")
			needsDownloadBackupFromS3 = true
		} else {
			isDataEmpty, err := isDataEmpty("block_info")
			if err != nil {
				log.Error().Err(err).Msg("Failed to check if data is empty")
				return err
			} else if isDataEmpty {
				log.Info().Msg("Tables exist but are empty - restoring from S3")
				needsDownloadBackupFromS3 = true
			}
		}

		if needsDownloadBackupFromS3 {
			log.Info().Msg("DB is empty - restoring from S3")
			_, err := restoreBackupFromS3()
			if err != nil {
				log.Err(err).Msg("Failed restoring DB and start fetching blockchain data from scratch")
			} else {
				log.Info().Msg("Successfully loaded dump")
			}
		} else {
			log.Info().Msg("DB is not empty - skipping restore from S3")
		}

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
			generateBlocksLoop(ctx, signalChan, heightsChan, exitWhenCaughtUp, bootstrapBlockHeight)
		}

		log.Info().Msg("Exited main loop, waiting for subroutines to finish...")
		cancel()
	}
	return nil
}

func getStartingHeight(bootstrapBlockHeight int64) (uint64, error) {
	if bootstrapBlockHeight > 0 {
		log.Info().Msgf("Starting synchronization from block height: %d", bootstrapBlockHeight)
		return uint64(bootstrapBlockHeight), nil // Return BLOCKHEIGHT as starting height
	}

	// If BLOCKHEIGHT is not set, get the last processed height from the database
	lastProcessedHeight, err := getLatestBlockHeightFromDB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to getLatestBlockHeightFromDB")
		return 0, err // Return 0 and the error if fetching fails
	}
	return lastProcessedHeight, nil
}

// Generates the block heights to process in an infinite loop
func generateBlocksLoop(ctx context.Context, signalChan <-chan os.Signal, heightsChan chan<- uint64, exitWhenCaughtUp bool, bootstrapBlockHeight int64) {
	for {
		// Get the starting height
		startingHeight, err := getStartingHeight(bootstrapBlockHeight)
		if err != nil {
			log.Error().Err(err).Msg("Error getting starting height, exiting...")
			return // Exit if there's an error
		}
		chainLatestHeight, err := getLatestHeight()
		if err != nil {
			log.Error().Err(err).Msg("Failed to getLatestHeight from chain")
		}
		log.Info().Msgf("Processing heights from %d to %d", startingHeight, chainLatestHeight)
		// Emit heights to process into channel
		for w := startingHeight; w <= chainLatestHeight; w++ {
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
				continue
			}
			log.Info().Msgf("fetchBlock height: %d, len(TXs): %d", height, len(block.Data.Txs))
			err = writeBlock(config, block)
			if err != nil {
				log.Error().Err(err).Msgf("Worker: Failed to writeBlock, height: %d", height)
				continue
			}

			log.Info().Msgf("Write height: %d", height)

			if len(block.Data.Txs) > 0 {
				log.Info().Msgf("Processing txs at height: %d", height)
				wgTxs := sync.WaitGroup{}
				txSemaphore := make(chan struct{}, maxConcurrentTxPerRoutine)
				for _, encTx := range block.Data.Txs {
					txSemaphore <- struct{}{} // Acquire a token
					wgTxs.Add(1)
					go func(encTx string) {
						defer wgTxs.Done()
						defer func() { <-txSemaphore }()             // Release the token
						err := processTx(ctx, &wgTxs, height, encTx) // Pass context and wait group
						if err != nil {
							log.Error().Err(err).Msgf("Failed to process transaction at height: %d", height)
						}
					}(encTx)
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
