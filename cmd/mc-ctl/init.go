package main

import "fmt"

func (a app) init() error {
	if err := a.ensureRuntimeLayout(); err != nil {
		return err
	}
	if err := a.checkRuntimeOwnership(); err != nil {
		return err
	}
	if err := a.ensureComposeEnv(); err != nil {
		return err
	}
	if err := a.ensureSecrets(); err != nil {
		return err
	}
	if err := a.renderWorldPaperGlobal(); err != nil {
		return err
	}
	if err := a.renderLimboConfig(); err != nil {
		return err
	}
	fmt.Printf("Initialized runtime and secrets under %s\n", a.baseDir)
	return nil
}
