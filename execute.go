package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/allora-network/allora-indexer/types"
	"github.com/rs/zerolog/log"
)

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

func DecodeTx(config ClientConfig, params string, blockHeight uint64) (types.Tx, error) {
	var result types.Tx

	// Determine the appropriate version based on block height
	var alloradPath string
	switch {
	case blockHeight >= 1574267:
		alloradPath = "/usr/local/bin/allorad" // 0.6.0
	case blockHeight >= 1296200:
		alloradPath = "/usr/local/bin/previous/v5/allorad" // 0.5.0
	case blockHeight >= 1004550:
		alloradPath = "/usr/local/bin/previous/v4/allorad" // 0.4.0
	case blockHeight >= 812000:
		alloradPath = "/usr/local/bin/previous/v3/allorad" // 0.3.0
	default:
		alloradPath = "/usr/local/bin/previous/v2/allorad" // 0.2.14
	}

	// Update the config to use the selected allorad binary
	config.CliApp = alloradPath

	result, err := ExecuteCommandByKey[types.Tx](config, "decodeTx", params)
	if err == nil {
		return result, nil
	}

	return result, err
}
