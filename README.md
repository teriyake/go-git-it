# go-git-it (ggi)

`go-git-it` (alias `ggi`) is a cli to-do app that leverages Github apis to manage tasks, set deadlines, and invite others to collaborate on your agenda.

## Installation

To build from source, clone this repo first:

```bash
git clone https://github.com/teriyake/go-git-it.git && cd go-git-it
```
Then, run the following to build:
```bash
go build
```
Optionally, you can add it to your path:
```bash
export PATH=$PATH:/path/to/your/install/directory/go-git-it
```

## Usage
```
./ggi [command]
```
> [!TIP]  
> If you are a first-time user, it is recommended to authenticate with Github first by executing `./ggi login`.  

> [!IMPORTANT]  
> You must manually **install** the app during the device flow authentication (otherwise api calls won't work). [screenshot goes here]

Aliases: 
- `ggi`
- `gg-it`
- `go-git-it`  

Available Commands:  
- `add`: add a new task  
- `choose-repo`: choose an existing to-do repo to work with  
- `del-repo`: delete an existing to-do repo  
- `del-task`: delete a task file in the current to-do repo  
- `done`: mark a to-do item as done by closing the corresponding Github issue  
- `help`: help about any command  
- `info`: info on current user  
- `login`: set up Git credentials  
- `mark`: mark a to-do item with a status  
- `new-repo`: create a new to-do repo  
- `whoami`: verify your Git auth status  

Flags:
- `-h`, `--help`: help for ggi

Use `./ggi [command] --help` for more information about a command.


## Contributing

Pull requests are welcome!  
Please open an issue if you encounter any bugs 🐛🐜🪲

## Author
[Teri Ke](https://github.com/teriyake)

## License

[MIT](https://choosealicense.com/licenses/mit/)
