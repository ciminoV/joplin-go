# Golang CLI for the Joplin Data API
`Joplingo-cli` is a command line interface to the notes stored in your Joplin instance.
It implements the bare minimum of the Joplin API. For the time being, it only allows you to create/edit and retrieve the notes from the default notebook.

## Authorisation
If you use the Joplin destkop version, the first time you run `Joplingo-cli`, it will try to get an authorisation token from your running local Joplin instance.
Switching to your local Joplin instance, there is a dialog asking you to grant or deny access to your data. Granting access will return the authorisation token back to `Joplingo-cli` and stored in a file called `.joplin-auth-token` in your home directory.

If you use the Joplin terminal version on Linux, you will need to manually create the `.joplin-auth-token` file in your home directory, with the following command:
```shell
cat ~/.config/joplin/settings.json | jq -r '."api.token"' > ~/.joplin-auth-token
```

## Commands

### Help
Calling `Joplingo-cli` with the command line option `-h` or `--help` will provide you with a brief description of the available commands and options:
```shell
Usage:
  joplingo-cli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  createnote  Create a new note from an existing file. Optionally specify the format.
  deletenote  Delete a note.
  getallnotes Retrieve all the notes. Optionally specify which fields and in which order.
  getnote     Retrieve a note with a given ID. Optionally specify which fields to return.
  help        Help about any command
  updatenote  Update a note with a given ID. Specify which fields to update with the corresponding values.

Flags:
  -h, --help      help for joplingo-cli
  -v, --verbose   verbose output

Use "joplingo-cli [command] --help" for more information about a command.
```

### Useful commands
Retrieve all the notes:
```shell
joplingo-cli getallnotes id title -f title -d asc
```

Get the body of a single note:
```shell
joplingo-cli getnote <id> body
```

Create a new note from a local file and delete it afterwards:
```shell
joplingo-cli createnote ~/mynewnote.md -f markdown -d 
```

Update the title of an existing note:
```shell
joplingo-cli updatenote <id> title newtitle
```

Delete permanently a note:
```shell
joplingo-cli deletenote <id> -p
```

## Example usage
If you use the terminal version on Linux, the Joplin server can be run on start, by adding the following lines to you `.xinitrc`:
```shell
joplin server start >/dev/null 2>&1 &
disown
```

Moreover, it can be used together with `dmenu` prompt to show all the remote notes and open a specific one inside `neovim`:
```shell
# Process the JSON piped input and populate an associative array with keys: id and values: title id
# (The spacing between title and id is for output purposes only)
declare -A items
json_input=$(echo $(joplingo-cli getallnotes id title -f title -d asc))
while IFS= read -r item; do
    id=$(echo "$item" | jq -r '.id')
    title=$(echo "$item" | jq -r '.title')
    items["$id"]="$(printf "%-*s %s" "30" "$title" "$id")"
done < <(echo "$json_input" | jq -c '.[]')

# Select a note from title and id
selected_item=$(printf "%s\n" "${items[@]}" | dmenu -i -p "Open note" -l 10)

# Check if a note is selected
if [[ -n $selected_item ]]; then
    selected_id="${selected_item##* }" # get only the id part of the value
    note_body=$(joplingo-cli getnote "$selected_id" body | jq -r '.body' | sed 's/\\n/\n/g') # pretty format the body
    
    # Create a temporary file with the note body and open it in nvim
    temp_file=$(mktemp /tmp/joplinote-XXXXXX --suffix=.md)
    echo -e "$note_body" > "$temp_file"
    st -e nvim -p "$temp_file"
    
    # Update remote note body
    joplingo-cli updatenote "$selected_id" body "$(cat $temp_file)"    
    
    # Clean up the temporary file
    trap 'rm -f "$temp_file"' EXIT
fi
```

## TO-DO
- [ ] Allows you to manage notebooks, tags and resources
- [ ] Add a search command
- [ ] Retrieve the auth token programmatically also with Joplin terminal
