package command

import "io"

type GlobalOptions struct {
	// Version shows the command's version
	Version bool `name:"version" help:"Show version"`

	// Stdin is os.Stdin by default. Useful for test.
	// This option works but not recommend to use for general usage.
	Stdin io.Reader `name:"interface-stdin" type:"stdin" default:"-" hidden:""`

	// Stdout is os.Stdout by default. Useful for test.
	// This option works but not recommend to use for general usage.
	Stdout io.Writer `name:"interface-stdout" type:"stdout" default:"-" hidden:""`

	// Stderr is os.Stderr by default. Useful for test.
	// This option works but not recommend to use for general usage.
	Stderr io.Writer `name:"interface-stderr" type:"stderr" default:"-" hidden:""`
}
