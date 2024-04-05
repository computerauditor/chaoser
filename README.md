# chaoser

A simple go script to extract subdomains from https://chaos.projectdiscovery.io/ and save into output file or date-wise

# Install
```
go install github.com/computerauditor/chaoser@latest
```

# Usage

```
chaoser [options]
```

# Options
```
-c: Number of concurrent download threads (default 30)
-o: The name and location of the output file
```

# Example
```
chaoser -c 60 -o /path/to/my_output.txt
```

# Credit
1) Mee
2) @rudSarkar

Link to the origional project:-

```
https://github.com/rudSarkar/chaosextract
```
