# chaosantor

A simple go script to extract subdomains from https://chaos.projectdiscovery.io/ and save into output file.

# Install
```
go install github.com/computerauditor/chaosantor/@latest
```

# Usage

```
chaosantor [options]
```

# Options
```
-c: Number of concurrent download threads (default 30)
-o: The name and location of the output file
```

# Example
```
chaosantor -c 60 -o /path/to/my_output.txt
```
