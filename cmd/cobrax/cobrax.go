package cobrax

import "github.com/spf13/cobra"

// Hook is a type that represents a function that takes a cobra command and its arguments.
type Hook func(cmd *cobra.Command, args []string)

// HookE is a type that represents a function that takes a cobra command and its arguments and returns an error.
type HookE func(cmd *cobra.Command, args []string) error

// Hooks is a function that takes a variable number of Hook functions and returns a new Hook function.
// The returned Hook function, when called, will execute all the provided Hook functions in order.
func Hooks(handlers ...Hook) Hook {
	return func(cmd *cobra.Command, args []string) {
		for _, handler := range handlers {
			handler(cmd, args)
		}
	}
}

// HooksE is a function that takes a variable number of HookE functions and returns a new HookE function.
// The returned HookE function, when called, will execute all the provided HookE functions in order.
// If any of the HookE functions return an error, the execution is stopped and the error is returned.
func HooksE(handlers ...HookE) HookE {
	return func(cmd *cobra.Command, args []string) error {
		for _, handler := range handlers {
			if err := handler(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
