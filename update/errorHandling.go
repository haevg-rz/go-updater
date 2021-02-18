package updater

import "fmt"

func printErrors(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
