# Command-Line

Package main provides a command-line interface for executing shortcuts defined in a JSON file.

## Usage

- Initializes by loading a JSON file specified in the `SHORTCUT` environment variable or in the home or current directory, named "shortcut.json".
- It parses the command-line arguments, retrieves the shortcut specified by the first argument, and executes it with the remaining arguments.

The program is invoked with the name of the shortcut to execute followed by its arguments, if any.
If the shortcut is not found, an error message is printed and the program exits with a non-zero status code.

## Example

shortcut.json 
```json
{
  "ssh": {
    "name": "ssh",
    "args": [
      "root@%s"
    ]
  }
}
```

run
```bash
sc ssh 127.0.0.1
```
equal run
```bash
ssh root@127.0.0.1
```