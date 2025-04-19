package ouidb

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type OUIEntry struct {
	Prefix       string
	Manufacturer string
}

const (
	dataDir       = "./data"
	savedFileName = "manuf.db"
)

var ouiData map[string]string

func downloadManufDB() error {
	fmt.Println("Fetching latest OUI database from Wireshark...")
	url := "https://www.wireshark.org/download/automated/data/manuf.gz"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading the OUI database:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to download file: status code %d\n", resp.StatusCode)
		return fmt.Errorf("Failed to download file: status code %d", resp.StatusCode)
	}

	// Create target directory if it doesn't exist
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		fmt.Println("Failed to create output directory:", err)
		return err
	}

	// Open gzip reader
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		fmt.Println("Error creating GZIP reader:", err)
		return err
	}
	defer gzReader.Close()

	// Create output file
	outputPath := filepath.Join(dataDir, savedFileName)
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return err
	}
	defer outFile.Close()

	// Copy uncompressed data to the file
	_, err = io.Copy(outFile, gzReader)
	if err != nil {
		fmt.Println("Error writing decompressed data to file:", err)
		return err
	}

	fmt.Println("OUI database updated and saved to", outputPath)
	return nil
}

func normalizeMACPrefix(prefix string) string {
	if strings.Contains(prefix, "/") {
		parts := strings.Split(prefix, "/")
		mac := parts[0]
		bits := parts[1]

		macParts := strings.Split(mac, ":")
		switch bits {
		case "28":
			// e.g., FC:D2:B6:30/28      CoetCostruzi    Coet Costruzioni Elettrotecniche
			return fmt.Sprintf("%s:%s", strings.Join(macParts[:3], ":"), string(macParts[3][0]))
		case "36":
			// e.g., 8C:1F:64:DC:70/36   WideSwathRes    Wide Swath Research, LLC
			return fmt.Sprintf("%s:%s", strings.Join(macParts[:4], ":"), string(macParts[4][0]))
		default:
			return mac // unknown bits, just return as-is
		}
	}
	return prefix
}

func UpdateDatabase(ouiFile string) error {
	file, err := os.Open(ouiFile)
	if err != nil {
		return fmt.Errorf("failed to open manuf file: %w", err)
	}
	defer file.Close()

	ouiData = make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue // skip comments and empty lines
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // skip malformed lines
		}
		macPrefix := parts[0]
		shortName := parts[1]
		normPrefix := normalizeMACPrefix(macPrefix)
		ouiData[normPrefix] = shortName
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fmt.Println("OUI database updated successfully with", len(ouiData), "entries.")
	return nil
}

// Lookup try the most specific (longest) match
func Lookup(mac string) (string, error) {
	ouiFile := filepath.Join(dataDir, savedFileName)
	if _, err := os.Open(ouiFile); err == nil {
		//fmt.Println("File already exists. No need to download again")
	} else {
		if err := downloadManufDB(); err == nil {
			return "", fmt.Errorf("Error in doenloading data file. err = %v", err)
		}
	}

	// Update the cache with manuf DB file
	if err := UpdateDatabase(ouiFile); err != nil {
		return "", err
	}

	// Wireshark manufacturer database uses all caps
	mac = strings.ToUpper(mac)
	parts := strings.Split(mac, ":")

	// Filter out any empty parts (e.g. due to trailing colon)
	filtered := []string{}
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	parts = filtered

	switch {
	case len(parts) >= 5:
		// Try starting from 36-bit match: 4 octets + 1 nibble → 00:00:20:33:1
		key36 := fmt.Sprintf("%s:%s", strings.Join(parts[:4], ":"), parts[4][:1])
		if val, ok := ouiData[key36]; ok {
			return val, nil
		}

		key28 := fmt.Sprintf("%s:%s", strings.Join(parts[:3], ":"), parts[3][:1])
		if val, ok := ouiData[key28]; ok {
			return val, nil
		}

		key24 := strings.Join(parts[:3], ":")
		if val, ok := ouiData[key24]; ok {
			return val, nil
		}

	case len(parts) >= 4:
		// Try starting from 28-bit match: 3 octets + 1 nibble → 00:00:20:3
		key28 := fmt.Sprintf("%s:%s", strings.Join(parts[:3], ":"), parts[3][:1])
		if val, ok := ouiData[key28]; ok {
			return val, nil
		}

		key24 := strings.Join(parts[:3], ":")
		if val, ok := ouiData[key24]; ok {
			return val, nil
		}

	case len(parts) >= 3:
		// Try 24-bit match: 3 octets → 00:00:20
		key24 := strings.Join(parts[:3], ":")
		if val, ok := ouiData[key24]; ok {
			return val, nil
		}
	}
	return "", errors.New("Manufacturer not found in any block.")
}
