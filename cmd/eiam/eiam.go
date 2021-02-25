package main

import (
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func main() {
	cmd := cmd.NewEphemeralIamCommand()
	eiamutil.CheckError(cmd.Execute())

	// f, err = os.Create("mem.prof")
	// if err != nil {
	// 	log.Fatal("could not create memory profile: ", err)
	// }
	// defer f.Close() // error handling omitted for example
	// runtime.GC()    // get up-to-date statistics
	// if err := pprof.WriteHeapProfile(f); err != nil {
	// 	log.Fatal("could not write memory profile: ", err)
	// }
}
