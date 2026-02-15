package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func executeInstall(command string, sudoFlag bool, useSudo string, alwaysRun bool) error {
	shouldSudo := sudoFlag

	if !sudoFlag && useSudo == "ask" {
		fmt.Print("Use sudo? [y/N] ")
		var confirm string
		if _, err := fmt.Scanln(&confirm); err != nil {
			shouldSudo = false
		} else {
			shouldSudo = confirm == "y" || confirm == "Y"
		}
	} else if !sudoFlag && useSudo == "true" {
		shouldSudo = true
	}

	if shouldSudo {
		command = "sudo " + command
	}

	if !alwaysRun {
		fmt.Print("Execute this command? [y/N] ")
		var confirm string
		if _, err := fmt.Scanln(&confirm); err != nil {
			confirm = ""
		}

		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborted.")

			return nil
		}
	}

	fmt.Printf("\nExecuting: %s\n\n", command)

	cmd := exec.CommandContext(context.Background(), "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func promptBatchConfig(
	sudoFlag, sudoOnlySystemFlag, skipIfBlindFlag, miseEnabled bool,
	reuseConfig string, alwaysRun bool,
) *BatchConfig {
	var batchCfg *BatchConfig
	shouldReuse := reuseConfig == "true"
	shouldAsk := reuseConfig == "ask"

	if shouldAsk {
		fmt.Print("Use same configuration for all tools? [Y/n] ")
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		shouldReuse = confirm != "n" && confirm != "N"
	}

	if shouldReuse {
		batchCfg = &BatchConfig{
			UseSudo:        sudoFlag,
			SudoOnlySystem: sudoOnlySystemFlag,
			SkipIfBlind:    skipIfBlindFlag,
			UseMise:        miseEnabled,
			AlwaysRun:      alwaysRun,
		}

		if !sudoFlag && !sudoOnlySystemFlag {
			fmt.Print("Use sudo? [y/N/s=system-only] ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			switch confirm {
			case "y", "Y":
				batchCfg.UseSudo = true
			case "s", "S":
				batchCfg.SudoOnlySystem = true
			}
		}

		if !skipIfBlindFlag {
			fmt.Print("Skip tools without install method? [y/N] ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			batchCfg.SkipIfBlind = confirm == "y" || confirm == "Y"
		}

		if !miseEnabled {
			fmt.Print("Use mise for installations? [y/N] ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			batchCfg.UseMise = confirm == "y" || confirm == "Y"
		}
	}

	return batchCfg
}

func promptExecute() error {
	fmt.Print("Execute? [y/N] ")
	var confirm string
	_, _ = fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		return fmt.Errorf("skipped: user declined")
	}

	return nil
}
