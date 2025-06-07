### chaoser

`chaoser` fetches ProjectDiscovery’s Chaos subdomain datasets and either aggregates them into a single file or “decompiles” them into per-program folders.

### Install
```
go install github.com/computerauditor/chaoser@latest
```

**(Optional)** Move it into your `$PATH`:

   ```bash
   mv chaoser /usr/local/bin/
   ```

### Usage

```
chaoser [options]
```

### Example
```
chaoser -c 60 -o /path/to/my_output.txt
```
### Detailed Documentation

```
[INFO] Usage of chaoser:
  -c, --concurrent int       number of concurrent downloads (default 30)
  -o, --output string        output file or directory (default: chaos-output-<date>)
  -t, --target string        fetch only programs/domains matching this substring
  -b, --bounty-only          fetch only paid-bounty programs
  -s, --swag-only            fetch only swag-reward programs
  -a, --all                  fetch all program types (default)
  -sp, --show-programs       list all programs in index.json and exit
  -d, --decompile            extract each program into its own folder under output dir
  -v, --verbose              verbose logging (debug)
```

> **Note:** Flags `-b/--bounty-only`, `-s/--swag-only`, and `-a/--all` are mutually exclusive. If you specify `-b` or `-s`, `-a` is automatically disabled.

---

### ⚙️ Flags

| Short | Long              | Type     | Default                     | Description                                                                        |
| ----- | ----------------- | -------- | --------------------------- | ---------------------------------------------------------------------------------- |
| `-c`  | `--concurrent`    | `int`    | `30`                        | Max concurrent downloads (controls goroutine semaphore).                           |
| `-o`  | `--output`        | `string` | `chaos-output-<YYYY-MM-DD>` | Path to output **file** (no `-d`) or **directory** (with `-d`).                    |
| `-t`  | `--target`        | `string` | `""`                        | Substring filter on program name or URL (case-insensitive).                        |
| `-b`  | `--bounty-only`   | `bool`   | `false`                     | Include **only** programs with `"bounty": true`.                                   |
| `-s`  | `--swag-only`     | `bool`   | `false`                     | Include **only** programs with `"swag": true`.                                     |
| `-a`  | `--all`           | `bool`   | `true`                      | Include **all** `"bounty"` and `"swag"` programs.                                  |
| `-sp` | `--show-programs` | `bool`   | `false`                     | Print list of **all** programs (with URL & type) and exit.                         |
| `-d`  | `--decompile`     | `bool`   | `false`                     | Extract each program’s ZIP into its own subfolder under the base output directory. |
| `-v`  | `--verbose`       | `bool`   | `false`                     | Enable **DEBUG-level** logging (each HTTP GET, unzip, etc.).                       |

---

### 📂 Output Modes

1. **Single-file (default)**

   * Creates `chaos-output-<YYYY-MM-DD>.txt` (or your `-o` value + `.txt`).
   * Appends all subdomain lists, one after another.

2. **Decompile mode** (`-d`)

   * Creates a base directory `chaos-output-<YYYY-MM-DD>` (or your `-o` value).
   * Under it, one subfolder per program (sanitized name).
   * Extracts each ZIP’s internal files directly into that program’s folder.

---

### 🚀 Examples

#### 1. Show all programs

```bash
$ ./chaoser --show-programs
[INFO] Programs available in Chaos:
 - Epic Games                  | https://chaos-data.projectdiscovery.io/epic_games.zip    | bounty
 - OnePlus                     | https://chaos-data.projectdiscovery.io/oneplus.zip       | bounty
 - Government of Netherlands   | https://chaos-data.projectdiscovery.io/starbucks.com.zip | swag
 ... (etc)
```

#### 2. Fetch all “swag” programs, single file

```bash
$ ./chaoser -s
[INFO] Starting fetch with concurrency=30
[INFO] Filtering: swag-only
[INFO] Total programs in index: 120
[INFO] 12 programs to fetch
[INFO] Writing all results to chaos-output-2025-06-08.txt
[INFO] Downloaded & appended https://chaos-data.projectdiscovery.io/starbucks.com.zip
[INFO] Downloaded & appended https://chaos-data.projectdiscovery.io/nasa.gov.zip
... 
[INFO] All done! 🎉
```

Result:

```
$ ls
chaoser  chaos-output-2025-06-08.txt

$ head chaos-output-2025-06-08.txt
sub1.starbucks.com
sub2.starbucks.com
...
sub1.nasa.gov
sub2.nasa.gov
...
```

#### 3. Fetch only “Adobe” (bounty) with verbose logs

```bash
$ ./chaoser -b -t adobe -v
[INFO] Starting fetch with concurrency=30
[INFO] Filtering: bounty-only
[INFO] Filtering programs containing substring: "adobe"
[INFO] Total programs in index: 120
[INFO] 1 programs to fetch
[INFO] Writing all results to chaos-output-2025-06-08.txt
[DEBUG] GET https://chaos-data.projectdiscovery.io/adobe.zip
[INFO] Downloaded & appended https://chaos-data.projectdiscovery.io/adobe.zip
[INFO] All done! 🎉
```

#### 4. Decompile all bounty programs into per-folder output

```bash
$ ./chaoser -b -d
[INFO] Starting fetch with concurrency=30
[INFO] Filtering: bounty-only
[INFO] Total programs in index: 120
[INFO] 108 programs to fetch
[INFO] Decompiling into directory chaos-output-2025-06-08/
[INFO] Processed Adobe
[INFO] Processed GitHub
...
[INFO] All done! 🎉
```

Result:

```
$ tree chaos-output-2025-06-08
chaos-output-2025-06-08
├── Adobe
│   └── adobe.com.txt
├── GitHub
│   └── github.com.txt
└── ... (more program folders)
```

---

### 🚧 Troubleshooting

* **`flag provided but not defined`**
  Make sure you’re running the **latest** binary (rebuild after edits).

* **Slow downloads / timeouts**

  * Reduce `-c` concurrency.
  * Add retry logic manually to `http.Get` calls if needed.


### Credit
1) Mee
2) @rudSarkar

Link to the origional project from which I took inspiration from :-

```
https://github.com/rudSarkar/chaosextract
```
