// Command ttanic manages archive folders: compressing, decompressing, and
// browsing directories as .tar.zst archives.
package main

import "fmt"

// version is stamped by the release pipeline via -ldflags; "dev" otherwise.
var version = "dev"

func main() {
	fmt.Println("ttanic " + version)
}
