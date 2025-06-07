package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProgramEntry mirrors each object in index.json
type ProgramEntry struct {
	URL     string `json:"URL"`
	Program string `json:"program"`
	Bounty  bool   `json:"bounty"`
	Swag    bool   `json:"swag"`
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		`[INFO] Usage of chaoser:
  -c, --concurrent int       number of concurrent downloads (default 30)
  -o, --output string        output file or directory (default: chaos-output-<date>)
  -t, --target string        fetch only programs/domains matching this substring
  -b, --bounty-only          fetch only paid-bounty programs
  -s, --swag-only            fetch only swag-reward programs
  -a, --all                  fetch all program types (default)
  -sp, --show-programs       list all programs in index.json and exit
  -d, --decompile            extract each program into its own folder under output dir
  -v, --verbose              verbose logging (debug)
`)
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}

func main() {
	// Flags
	concurrent := flag.Int("c", 30, "number of concurrent downloads")
	flag.IntVar(concurrent, "concurrent", 30, "number of concurrent downloads")

	output := flag.String("o", "", "output file or directory (default: chaos-output-<date>)")
	flag.StringVar(output, "output", "", "output file or directory (default: chaos-output-<date>)")

	target := flag.String("t", "", "fetch only programs/domains matching this substring")
	flag.StringVar(target, "target", "", "fetch only programs/domains matching this substring")

	bountyOnly := flag.Bool("b", false, "bounty-only: fetch only paid-bounty programs")
	flag.BoolVar(bountyOnly, "bounty-only", false, "bounty-only: fetch only paid-bounty programs")

	swagOnly := flag.Bool("s", false, "swag-only: fetch only swag-reward programs")
	flag.BoolVar(swagOnly, "swag-only", false, "swag-only: fetch only swag-reward programs")

	allTypes := flag.Bool("a", true, "fetch all program types (default)")
	flag.BoolVar(allTypes, "all", true, "fetch all program types (default)")

	showPrograms := flag.Bool("sp", false, "list all programs in index.json and exit")
	flag.BoolVar(showPrograms, "show-programs", false, "list all programs in index.json and exit")

	decompile := flag.Bool("d", false, "decompile: extract each program into its own folder under output dir")
	flag.BoolVar(decompile, "decompile", false, "decompile: extract each program into its own folder under output dir")

	verbose := flag.Bool("v", false, "verbose output (debug logging)")
	flag.BoolVar(verbose, "verbose", false, "verbose output (debug logging)")

	flag.Usage = usage
	flag.Parse()

	// Fetch & parse index.json
	idxURL := "https://chaos-data.projectdiscovery.io/index.json"
	if *verbose {
		log.Printf("[DEBUG] GET %s", idxURL)
	}
	resp, err := http.Get(idxURL)
	if err != nil {
		log.Fatalf("[ERROR] GET index.json: %v", err)
	}
	defer resp.Body.Close()

	var entries []ProgramEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		log.Fatalf("[ERROR] decode index.json: %v", err)
	}

	// Show programs and exit
	if *showPrograms {
		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Program) < strings.ToLower(entries[j].Program)
		})
		fmt.Println("[INFO] Programs available in Chaos:")
		for _, e := range entries {
			types := []string{}
			if e.Bounty {
				types = append(types, "bounty")
			}
			if e.Swag {
				types = append(types, "swag")
			}
			fmt.Printf(" - %-20s | %-50s | %s\n",
				e.Program, e.URL, strings.Join(types, ","))
		}
		return
	}

	// Validate mutually exclusive flags
	if *bountyOnly && *swagOnly {
		log.Fatalf("[ERROR] -bounty-only and -swag-only cannot be used together")
	}
	if *bountyOnly || *swagOnly {
		*allTypes = false
	}
	if !*allTypes && !*bountyOnly && !*swagOnly {
		log.Fatalf("[ERROR] must specify at least one of -all, -bounty-only, or -swag-only")
	}

	// Prepare filters
	includeBounty := *allTypes || *bountyOnly
	includeSwag := *allTypes || *swagOnly
	targetSubstr := strings.ToLower(*target)

	log.Printf("[INFO] Starting fetch with concurrency=%d", *concurrent)
	if *bountyOnly {
		log.Printf("[INFO] Filtering: bounty-only")
	}
	if *swagOnly {
		log.Printf("[INFO] Filtering: swag-only")
	}
	if *target != "" {
		log.Printf("[INFO] Filtering programs containing substring: %q", *target)
	}
	log.Printf("[INFO] Total programs in index: %d", len(entries))

	// Build list of entries to download
	var toDownload []ProgramEntry
	for _, e := range entries {
		if e.Bounty && !includeBounty {
			continue
		}
		if e.Swag && !includeSwag {
			continue
		}
		if !e.Bounty && !e.Swag {
			continue
		}
		if targetSubstr != "" {
			if !strings.Contains(strings.ToLower(e.Program), targetSubstr) &&
				!strings.Contains(strings.ToLower(e.URL), targetSubstr) {
				continue
			}
		}
		toDownload = append(toDownload, e)
	}

	if len(toDownload) == 0 {
		log.Printf("[INFO] No matching entries found. Exiting.")
		return
	}
	log.Printf("[INFO] %d programs to fetch", len(toDownload))

	// Determine output path
	dateStr := time.Now().Format("2006-01-02")
	baseOut := *output
	if baseOut == "" {
		baseOut = fmt.Sprintf("chaos-output-%s", dateStr)
	}

	// Single-file or directory setup
	var singleFH *os.File
	if !*decompile {
		fh, err := os.Create(baseOut + ".txt")
		if err != nil {
			log.Fatalf("[ERROR] cannot create %s.txt: %v", baseOut, err)
		}
		defer fh.Close()
		singleFH = fh
		log.Printf("[INFO] Writing all results to %s.txt", baseOut)
	} else {
		if err := os.MkdirAll(baseOut, 0o755); err != nil {
			log.Fatalf("[ERROR] cannot create directory %s: %v", baseOut, err)
		}
		log.Printf("[INFO] Decompiling into directory %s/", baseOut)
	}

	// Download & process concurrently
	var wg sync.WaitGroup
	sem := make(chan struct{}, *concurrent)

	for _, entry := range toDownload {
		wg.Add(1)
		sem <- struct{}{}

		go func(e ProgramEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			if *verbose {
				log.Printf("[DEBUG] GET %s", e.URL)
			}
			r, err := http.Get(e.URL)
			if err != nil {
				log.Printf("[ERROR] GET %s: %v", e.URL, err)
				return
			}
			defer r.Body.Close()

			data, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("[ERROR] read %s: %v", e.URL, err)
				return
			}

			zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				log.Printf("[ERROR] unzip %s: %v", e.URL, err)
				return
			}

			if !*decompile {
				// append all entries into single file
				for _, f := range zr.File {
					rc, err := f.Open()
					if err != nil {
						log.Printf("[ERROR] open entry: %v", err)
						continue
					}
					if _, err := io.Copy(singleFH, rc); err != nil {
						log.Printf("[ERROR] write single file: %v", err)
					}
					rc.Close()
				}
				log.Printf("[INFO] Appended %s", e.Program)
			} else {
				// extract into per-program folder
				dirName := sanitizeName(e.Program)
				progDir := filepath.Join(baseOut, dirName)
				if err := os.MkdirAll(progDir, 0o755); err != nil {
					log.Printf("[ERROR] mkdir %s: %v", progDir, err)
					return
				}

				fileCount := 0
				for _, f := range zr.File {
					rc, err := f.Open()
					if err != nil {
						log.Printf("[ERROR] open %s entry: %v", e.Program, err)
						continue
					}
					outPath := filepath.Join(progDir, f.Name)
					of, err := os.Create(outPath)
					if err != nil {
						log.Printf("[ERROR] create %s: %v", outPath, err)
						rc.Close()
						continue
					}
					if _, err := io.Copy(of, rc); err != nil {
						log.Printf("[ERROR] write %s: %v", outPath, err)
					}
					of.Close()
					rc.Close()
					fileCount++
				}
				log.Printf("[INFO] %s → extracted %d files into %s/", e.Program, fileCount, progDir)
			}
		}(entry)
	}

	wg.Wait()
	log.Printf("[INFO] All done! 🎉")
}
