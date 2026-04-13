package detectors

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Aadi-IRON/agni/config"
)

// Detects all unused constants present in the directory.
func DetectUnusedConstants(filePath string) {
	fmt.Println(config.CreateCompactBoxHeader("UNUSED CONSTANTS", config.BoldGreen))
	fmt.Println()
	fmt.Println(config.BoldYellow + "🔍 Detecting unused constants (Declared but not used):")
	if filePath == "" {
		fmt.Println("Please pass a valid directory path.")
		return
	}
	// Find unused constants
	unusedConsts, err := FindUnusedConsts(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Output results
	if len(unusedConsts) == 0 {
		fmt.Println()
		fmt.Println(config.BoldGreen + "✅  All constants in const.go are used in the project.")
		fmt.Println()
		return
	}
	fmt.Println()
	fmt.Println(config.BoldYellow + "Unused constants in const.go:-> ")
	for _, constant := range unusedConsts {
		fmt.Printf(config.Red+"- %s\n", constant)
	}
	fmt.Println()
}

// SearchConstantInProject checks if a constant is used in any `.go` file.
func SearchConstantInProject(projectDir, constant, excludeFile string) (bool, error) {
	found := false
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories; keep all .go files, including the const file.
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if filepath.Base(path) == excludeFile {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, constant) {
				found = true
				return nil // Found, stop scanning this file
			}
		}
		return nil
	})
	return found, err
}

// FindUnusedConstants finds constants in `const.go` that are not used anywhere in the project.
func FindUnusedConsts(filePath string) ([]string, error) {
	constants, err := ExtractConstsFromFile(filePath + "/config/const.go")
	excludeFile := "const.go"
	if err != nil {
		constants, err = ExtractConstsFromFile(filePath + "/config/Const.go")
		excludeFile = "Const.go"
		if err != nil {
			return nil, err
		}
	}
	var unusedConsts []string
	for _, constant := range constants {
		used, err := SearchConstantInProject(filePath, constant, excludeFile)
		if err != nil {
			return nil, err
		}
		if !used {
			unusedConsts = append(unusedConsts, constant)
		}
	}
	return unusedConsts, nil
}

// ExtractConstantsFromFile reads `const.go` and extracts constants defined in it.
func ExtractConstsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var constants []string
	scanner := bufio.NewScanner(file)
	constBlock := false // Track if inside a `const` block
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Detect start of a `const` block
		if strings.HasPrefix(line, "const (") {
			constBlock = true
			continue
		}
		// Detect end of a `const` block
		if constBlock && line == ")" {
			constBlock = false
			continue
		}
		// Extract constant names from the block
		if constBlock {
			// Assume constant names are the first word before the `=` sign
			parts := strings.Fields(line)
			if len(parts) > 0 {
				constants = append(constants, parts[0])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return constants, nil
}
