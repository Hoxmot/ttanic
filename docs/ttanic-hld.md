# ttanic HLD

## Description

`ttanic` (pronounced tee-tanic) is a short for tar-tanic. An app for managing your archive folders. It allows you to manage and automatically compress your directories.

## Need

Recently, I started working on organising and backing up my data. As a result, I realised that some of the data can be compressed to save some space. However, considering how long it takes to decompress. I'd like to manage it the compression of my files and directories more easily.

The problem is that I also don't remember all the commands to compress and decompress -- especially the more complicated ones. As a result, it's not easy to manage it manually. That's why I need something to help me with that.

## Features

I'd like the program to have to options to work with:

- CLI, where I directly give commands to do something
- TUI, which will be a bit easier and interactive way to use the app

I'd like to have a nice TUI, where the user can navigate tree structure. The user can navigate, select files, and directory, group compress, group decompress, etc.

The program also has a flow, where the user can select to compress all the files in the selected directory separately. The program also keeps the manifest of all the files in the compressed directories. As a result, the user can easily browse the contents of compressed archives.

The user can use the program in initialised directory as a tar-tanic project to keep the manifest and configs locally or it can be used as a simple compression tool standalone.

The program should have a `?` command to display the help menu.

The program should have `/` command for looking for particular files in the directory.

The user can use `<SUPER>/` to do fuzzy search from the files in the tar-tanic project or from the file from the directory the program was open down (if the program was opened outside of tar-tanic project).

I'd like to also have a config view for the user to modify the config. If the user is in the tar-tanic project, the program should edit directory-level settings. If the user outside of tar-tanic project, the user edits global settings.

All the TUI features are also available from CLI.

### User flows

#### Start

1. I open TUI in an uninitialised directory
2. I press `S` or type `:init` to start initialisation process
3. The programs launches a wizard to ask me questions about the project
4. After I answer to questions, the program creates manifest and config to initialise tar-tanic project.

#### Simple compression

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `c` for compression
4. The program compresses the directory/file

#### Simple decompression

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `i` for decompression (inflate)
4. The program compresses the directory/file

#### Deleting the directory

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `d` for deletion
	1. If there is a compressed directory/file the user wants to delete, the file/directory gets deleted
	2. If the program doesn't see the compressed directory/file in the same directory, it prompts for confirmation or compress and delete.
		1. If the I choose delete, the program deletes the directory/file
		2. If the I choose abort, the program aborts
		3. If the I choose compress and delete, the program compresses, then deletes the file/directory

#### Compres and delete

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `:` to start a command typing, and type `cd` for compression and deletion
4. The program compresses the directory/file and then deletes the original directory/file

#### Compress recursively

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `C` for recursive compression
4. The program compresses the all the files inside the selected directory into separate archives. If the selected thing is a file, it behaves like `c` (single compression)

#### Decompress recursively

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `I` for recursive decompression (inflate)
4. The program decompresses the all the archives inside the selected directory into separate archives. If the selected thing is a file, it behaves like `i` (single decompression)

#### Multi selection

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `v` for selecting the directory/file
4. I navigate to a different directory/file
5. I press `v` again to select the other directory/file
6. I press `V` for multi selection and can go up/down to select all the file I go thogh
7. I press `V` to stop multi-selection
8. I can type whatever command and it applies to all the files/directories I selected
9. I can press `ESC` to deselect all the files

## Tech-stack

I'd like to use `go` programming language with nice TUI made with tools from charm.land: `Bubble Tea`, `Huh`, and `Lip Gloss`.

For compression, I'd like to use (most probably) `zstd`. When I worked with bash, I used `tar`, but if there's library for that, I think it's better than invoking another program -- especially that the program might not be installed on host.

Of course, I'd like the program to be easily distributable/installable.

